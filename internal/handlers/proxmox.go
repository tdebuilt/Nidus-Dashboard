package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/proxmox"
)

// ProxmoxHandler handles Proxmox-related HTTP requests.
type ProxmoxHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

func (h *ProxmoxHandler) getProxmoxClient() (*proxmox.Client, error) {
	svc, err := h.DB.GetServiceByType("proxmox")
	if err != nil {
		return nil, err
	}
	if svc == nil {
		return nil, nil
	}

	log.Printf("proxmox: connecting to %s", svc.URL)
	client := proxmox.NewClient(svc.URL, nil)

	if svc.Credentials != "" {
		encKey, err := h.DB.GetSystemSetting("encryption_key")
		if err != nil || encKey == "" {
			log.Printf("proxmox: no encryption key found")
			return nil, err
		}
		creds, err := crypto.Decrypt(svc.Credentials, encKey)
		if err != nil {
			log.Printf("proxmox: failed to decrypt credentials: %v", err)
			return nil, err
		}
		var authData struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Token    string `json:"token"`
		}
		if err := json.Unmarshal([]byte(creds), &authData); err != nil {
			log.Printf("proxmox: credentials is not JSON, using as raw token")
			client.SetAPIToken(creds)
			return client, nil
		}
		switch {
		case authData.Token != "":
			log.Printf("proxmox: using API token (len=%d)", len(authData.Token))
			client.SetAPIToken(authData.Token)
		case authData.Username != "":
			log.Printf("proxmox: authenticating as %s", authData.Username)
			if err := client.Authenticate(authData.Username, authData.Password); err != nil {
				log.Printf("proxmox: auth failed: %v", err)
				return nil, err
			}
		default:
			log.Printf("proxmox: no token or username in credentials")
		}
	}

	return client, nil
}

// ListNodes godoc
// @Summary List Proxmox cluster nodes
// @Tags proxmox
// @Produce json
// @Success 200 {array} proxmox.NodeInfo
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /proxmox/nodes [get]
// @Security BearerAuth
func (h *ProxmoxHandler) ListNodes(w http.ResponseWriter, r *http.Request) {
	if val, ok := h.Cache.Get("proxmox:nodes"); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getProxmoxClient()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to connect to Proxmox"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Proxmox not configured"})
		return
	}

	nodes, err := client.ListNodes()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch nodes"})
		return
	}

	result := make([]proxmox.NodeInfo, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, proxmox.NodeInfo{
			Name:     n.Node,
			Status:   n.Status,
			CPUUsage: n.CPU,
			CPUCores: n.MaxCPU,
			MemUsed:  n.Mem,
			MemTotal: n.MaxMem,
			Uptime:   n.Uptime,
		})
	}

	h.Cache.Set("proxmox:nodes", result)
	writeJSON(w, http.StatusOK, result)
}

// ListVMs godoc
// @Summary List all VMs and LXC containers across nodes
// @Tags proxmox
// @Produce json
// @Success 200 {array} proxmox.VMInfo
// @Failure 502 {object} models.ErrorResponse
// @Router /proxmox/vms [get]
// @Security BearerAuth
func (h *ProxmoxHandler) ListVMs(w http.ResponseWriter, r *http.Request) {
	if val, ok := h.Cache.Get("proxmox:vms"); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getProxmoxClient()
	if err != nil {
		log.Printf("proxmox: getClient error: %v", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Proxmox not available"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Proxmox not available"})
		return
	}

	allVMs, err := client.ListAllVMs()
	if err != nil {
		log.Printf("proxmox: ListAllVMs error: %v", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch VMs"})
		return
	}

	result := make([]proxmox.VMInfo, 0, len(allVMs))
	for _, vm := range allVMs {
		result = append(result, proxmox.VMInfo{
			VMID:     vm.VMID,
			Name:     vm.Name,
			Node:     vm.Node,
			Type:     vm.Type,
			Status:   vm.Status,
			CPUUsage: vm.CPU,
			CPUCores: vm.CPUs,
			MemUsed:  vm.Mem,
			MemTotal: vm.MaxMem,
			Uptime:   vm.Uptime,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	h.Cache.Set("proxmox:vms", result)
	writeJSON(w, http.StatusOK, result)
}

// VMAction godoc
// @Summary Perform an action on a VM or LXC container
// @Tags proxmox
// @Produce json
// @Param node path string true "Node name"
// @Param vmType path string true "VM type" Enums(qemu, lxc)
// @Param vmid path int true "VM ID"
// @Param action path string true "Action to perform" Enums(start, stop, shutdown, reboot)
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /proxmox/vms/{node}/{vmType}/{vmid}/{action} [post]
// @Security BearerAuth
func (h *ProxmoxHandler) VMAction(w http.ResponseWriter, r *http.Request) {
	node := chi.URLParam(r, "node")
	vmType := chi.URLParam(r, "vmType")
	vmidStr := chi.URLParam(r, "vmid")
	action := chi.URLParam(r, "action")

	vmid, err := strconv.Atoi(vmidStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid VMID"})
		return
	}

	if vmType != "qemu" && vmType != "lxc" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid VM type"})
		return
	}

	validActions := map[string]bool{
		"start": true, "stop": true, "shutdown": true, "reboot": true,
	}
	if !validActions[action] {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid action"})
		return
	}

	client, err := h.getProxmoxClient()
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Proxmox not available"})
		return
	}

	taskID, err := client.VMAction(node, vmType, vmid, action)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "action failed: " + err.Error()})
		return
	}

	h.Cache.InvalidatePrefix("proxmox:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "task_id": taskID})
}
