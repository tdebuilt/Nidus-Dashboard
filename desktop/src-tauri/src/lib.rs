use std::sync::Mutex;
use tauri::Emitter;
use tauri::Manager;
use tauri::RunEvent;
use tauri_plugin_shell::process::{CommandChild, CommandEvent};
use tauri_plugin_shell::ShellExt;

struct SidecarState(Mutex<Option<CommandChild>>);

pub fn run() {
    let app = tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .manage(SidecarState(Mutex::new(None)))
        .setup(|app| {
            let handle = app.handle().clone();

            let sidecar = handle
                .shell()
                .sidecar("nidus")
                .expect("failed to create sidecar command")
                .args(["--desktop"]);

            let (mut rx, child) = sidecar.spawn().expect("failed to spawn sidecar");

            // Store the child handle for cleanup
            let state = handle.state::<SidecarState>();
            *state.0.lock().unwrap() = Some(child);

            // Read stdout to detect the port
            let handle_clone = handle.clone();
            tauri::async_runtime::spawn(async move {
                while let Some(event) = rx.recv().await {
                    match event {
                        CommandEvent::Stdout(line_bytes) => {
                            let line = String::from_utf8_lossy(&line_bytes);
                            let line = line.trim();

                            if let Some(port_str) = line.strip_prefix("NIDUS_PORT=") {
                                if let Ok(port) = port_str.parse::<u16>() {
                                    let url = format!("http://127.0.0.1:{}", port);

                                    // Wait briefly for the server to be ready
                                    tokio::time::sleep(std::time::Duration::from_millis(500))
                                        .await;

                                    if let Some(window) =
                                        handle_clone.get_webview_window("main")
                                    {
                                        let _ = window.navigate(url.parse().unwrap());
                                    }
                                }
                            }

                            let _ = handle_clone.emit("sidecar-stdout", line.to_string());
                        }
                        CommandEvent::Stderr(line_bytes) => {
                            let line = String::from_utf8_lossy(&line_bytes);
                            let _ = handle_clone.emit("sidecar-stderr", line.to_string());
                        }
                        CommandEvent::Terminated(status) => {
                            let _ =
                                handle_clone.emit("sidecar-terminated", format!("{:?}", status));
                            break;
                        }
                        _ => {}
                    }
                }
            });

            Ok(())
        })
        .build(tauri::generate_context!())
        .expect("error while building Nidus");

    app.run(|app_handle, event| {
        if let RunEvent::Exit = event {
            let state = app_handle.state::<SidecarState>();
            let child = state.0.lock().unwrap().take();
            if let Some(child) = child {
                let _ = child.kill();
            }
        }
    });
}
