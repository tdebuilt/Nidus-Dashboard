import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const DB_PATH = path.join(__dirname, '..', '..', 'data', 'e2e-test.db')
const AUTH_DIR = path.join(__dirname, '.auth')

export default function globalSetup() {
  // Clean up test database for a fresh run
  if (fs.existsSync(DB_PATH)) {
    fs.unlinkSync(DB_PATH)
  }
  // Also remove WAL/SHM files if present
  for (const suffix of ['-wal', '-shm']) {
    const p = DB_PATH + suffix
    if (fs.existsSync(p)) {
      fs.unlinkSync(p)
    }
  }

  // Clean up auth storage state
  if (fs.existsSync(AUTH_DIR)) {
    fs.rmSync(AUTH_DIR, { recursive: true })
  }
}
