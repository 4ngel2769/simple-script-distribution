# Admin API Documentation

## Authentication

All API endpoints require session-based authentication. Login first via the web interface at `/admin`.

## Endpoints

### Scripts Management

#### Get All Scripts
```http
GET /admin/scripts
```

**Response:**
```json
[
  {
    "name": "tor",
    "path": "tor", 
    "description": "Tor installation script",
    "icon": "🧅",
    "type": "local",
    "redirect_url": ""
  }
]
```

#### Create Script
```http
POST /admin/scripts
Content-Type: application/json

{
  "name": "script-name",
  "description": "Script description",
  "icon": "📜",
  "type": "local",
  "redirect_url": ""
}
```

#### Update Script
```http
PUT /admin/scripts/{name}
Content-Type: application/json

{
  "description": "Updated description",
  "icon": "🔧"
}
```

#### Delete Script
```http
DELETE /admin/scripts/{name}
```

### Script Content Management

#### Get Script Content
```http
GET /admin/scripts/{name}/content
```

**Response:**
```json
{
  "content": "#!/bin/bash\necho 'Hello World'"
}
```

#### Update Script Content
```http
PUT /admin/scripts/{name}/content
Content-Type: application/json

{
  "content": "#!/bin/bash\necho 'Updated script'"
}
```

### Index Page Management

#### Get Index Page Data
```http
GET /admin/index-page
```

#### Update Index Page
```http
POST /admin/index-page
Content-Type: application/json

{
  "scripts": [...]
}
```

## Error Responses

All endpoints return JSON error responses:

```json
{
  "error": "Error message description"
}
```

HTTP Status Codes:
- `200` - Success
- `400` - Bad Request
- `401` - Unauthorized
- `404` - Not Found
- `409` - Conflict (script already exists)
- `500` - Internal Server Error
