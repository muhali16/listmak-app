# API Documentation - ListMak Service

Base URL: `http://localhost:8080/api/v1` (Adjust port as necessary)

## Response Format

Semua response dari server akan memiliki format standar berikut:

```json
{
  "code": 200,                  // HTTP Status Code
  "success": true,              // Status request (true/false)
  "message": "Success message", // Pesan deskriptif
  "data": { ... },              // Data payload (bisa object/array/null)
  "request_id": "uuid...",      // ID unik untuk tracing req
  "latency": "12ms"             // Waktu proses server
}
```

## Authentication

### 1. Google Login

Menginisiasi proses login OAuth2 dengan Google.

- **Endpoint**: `/auth/google/login`
- **Method**: `GET`
- **Description**: Redirect user ke halaman login Google.

### 2. Google Callback

Callback yang dipanggil oleh Google setelah user berhasil login.

- **Endpoint**: `/auth/google/callback`
- **Method**: `GET`
- **Query Params**:
  - `code` (string, required): Authorization code dari Google.
- **Response**:
  - `Set-Cookie`: `X-User-Authentication-Token` (JWT)
  - `Body`:
    ```json
    {
      "code": 200,
      "success": true,
      "message": "Logged in successfully",
      "data": {
        "token": "eyJhbGcV...",
        "user": {
          "id": 1,
          "google_id": "12345",
          "email": "user@example.com",
          "name": "User Name",
          "avatar": "https://...",
          "role": "user",
          "created_at": "...",
          "updated_at": "..."
        }
      }
    }
    ```

### 3. Logout

Logout user dan menghapus cookie autentikasi.

- **Endpoint**: `/auth/logout`
- **Method**: `GET`
- **Headers**:
  - `Cookie`: `X-User-Authentication-Token=...`
- **Response**:
  ```json
  {
    "code": 200,
    "success": true,
    "message": "Logged out successfully",
    "data": null
  }
  ```

---

## Users

**Note**: Endpoint ini membutuhkan Autentikasi (Cookie / Middleware).

### 1. Get All Users

Mengambil daftar semua user.

- **Endpoint**: `/api/v1/users`
- **Method**: `GET`
- **Response**:
  ```json
  {
    "code": 200,
    "success": true,
    "message": "Success get users",
    "data": [
      {
         "id": 1,
         "google_id": "...",
         "email": "...",
         "name": "...",
         ...
      }
    ]
  }
  ```

### 2. Create User

Membuat user baru secara manual (selain via Google Login).

- **Endpoint**: `/api/v1/users`
- **Method**: `POST`
- **Request Body** (JSON):
  ```json
  {
    "google_id": "123456", // required
    "email": "new@mail.com", // required
    "name": "New User", // required
    "avatar": "http://...", // optional
    "role": "user" // optional, enum: 'user', 'admin'
  }
  ```
- **Response**:
  ```json
  {
    "code": 200,
    "success": true,
    "message": "Success create user",
    "data": {
      "id": 2,
      ...
    }
  }
  ```

---

## System Logs

### 1. Get All Logs

Melihat log aktivitas sistem dengan berbagai filter.

- **Endpoint**: `/api/v1/logs`
- **Method**: `GET`
- **Query Params** (Optional):
  - `request_id`: Filter by Request ID
  - `method`: Filter by HTTP Method (GET, POST, etc)
  - `path`: Filter by URL Path
  - `status_code`: Filter by HTTP Status Code
  - `client_ip`: Filter by User IP
- **Response**:
  ```json
  {
    "code": 200,
    "success": true,
    "message": "Success get logs",
    "data": [
      {
        "id": 10,
        "request_id": "uuid...",
        "method": "GET",
        "path": "/api/v1/users",
        "status_code": 200,
        "latency": "10ms",
        "client_ip": "127.0.0.1",
        "error_msg": "",
        "created_at": "..."
      }
    ]
  }
  ```

### 2. Get Log by Request ID

Melihat detail log berdasarkan Request ID spesifik.

- **Endpoint**: `/api/v1/logs/{request_id}`
- **Method**: `GET`
- **Path Params**:
  - `request_id` (string): ID unik request
- **Response**:
  ```json
  {
    "code": 200,
    "success": true,
    "message": "Success get log",
    "data": {
        "id": 10,
        "request_id": "uuid...",
        ...
    }
  }
  ```
