# Nidus API — curl Examples

## Authentication

### Login

```bash
curl -X POST http://localhost:3777/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"yourpassword"}'
```

Response:
```json
{"message":"login successful","user":{"id":1,"username":"admin","role":"admin"}}
```

The JWT token is set as a cookie (`nidus_token`). For curl, pass it with `-b`:

```bash
# Save cookie on login
curl -c cookies.txt -X POST http://localhost:3777/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"yourpassword"}'

# Use cookie for authenticated requests
curl -b cookies.txt http://localhost:3777/api/categories
```

Or use the Authorization header:

```bash
curl http://localhost:3777/api/categories \
  -H "Authorization: Bearer <your-jwt-token>"
```

### Check setup status

```bash
curl http://localhost:3777/api/auth/status
```

## Categories

### List categories

```bash
curl -b cookies.txt http://localhost:3777/api/categories
```

### Create a category

```bash
curl -b cookies.txt -X POST http://localhost:3777/api/categories \
  -H "Content-Type: application/json" \
  -d '{"name":"Media","icon":"tv"}'
```

## Widgets

### List widgets in a category

```bash
curl -b cookies.txt http://localhost:3777/api/categories/1/widgets
```

### Create a widget

```bash
curl -b cookies.txt -X POST http://localhost:3777/api/categories/1/widgets \
  -H "Content-Type: application/json" \
  -d '{"type":"docker","title":"My Docker","config":"{}","pos_x":0,"pos_y":0,"width":4,"height":0}'
```

## Services

### List configured services

```bash
curl -b cookies.txt http://localhost:3777/api/services
```

### Configure a service

```bash
curl -b cookies.txt -X PUT http://localhost:3777/api/services/portainer \
  -H "Content-Type: application/json" \
  -d '{"name":"Portainer","url":"https://portainer.local:9443","credentials":"{\"api_key\":\"your-api-key\"}"}'
```

### Test service connectivity

```bash
curl -b cookies.txt -X POST http://localhost:3777/api/services/portainer/test
```

## Docker

### List environments

```bash
curl -b cookies.txt http://localhost:3777/api/docker/environments
```

### List containers

```bash
curl -b cookies.txt http://localhost:3777/api/docker/environments/1/containers
```

### Container action (start/stop/restart)

```bash
curl -b cookies.txt -X POST http://localhost:3777/api/docker/environments/1/containers/abc123/restart
```

## Settings

### Get current settings

```bash
curl -b cookies.txt http://localhost:3777/api/settings
```

### Update settings

```bash
curl -b cookies.txt -X PUT http://localhost:3777/api/settings \
  -H "Content-Type: application/json" \
  -d '{"theme":"nord","language":"en","refresh_interval":15}'
```

## Config Export/Import

### Export as YAML

```bash
curl -b cookies.txt http://localhost:3777/api/config/yaml -o nidus-config.yaml
```

### Import YAML

```bash
curl -b cookies.txt -X POST http://localhost:3777/api/config/yaml \
  -H "Content-Type: application/json" \
  --data-binary @nidus-config.yaml
```
