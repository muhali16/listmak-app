# Listmak App

Full-stack collaborative ordering/list management app.

## Structure

```
listmak-app/
├── backend/    # Go REST API (Gin + GORM + PostgreSQL)
└── frontend/   # Vue 3 SPA (Vite)
```

---

# Backend

REST API Server untuk manajemen Listmak, dibangun dengan Go (Gin Framework), PostgreSQL, dan GORM.

## Tech Stack

- **Go** - Programming language
- **Gin** - Web framework
- **GORM** - ORM library
- **PostgreSQL** - Primary database
- **SQLite** - Request log storage (`logs.db`)
- **JWT** - Authentication
- **Swagger** - API documentation

## Requirements

- Go 1.21+
- PostgreSQL 14+
- Git

## Quick Start (Development)

### 1. Clone Repository

```bash
git clone https://github.com/muhali16/listmak-service.git
cd listmak-service
```

### 2. Setup Environment

```bash
cp .env.example .env
```

Edit `.env` sesuai konfigurasi lokal:

```env
PORT=9001
ENV=development
DEBUG=true

DB_HOST=localhost
DB_NAME=listmak_service
DB_PORT=5432
DB_USER=postgres
DB_PASS=your_password

GOOGLE_CLIENT_ID=your_id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your_secret
GOOGLE_REDIRECT_URL=http://localhost:9001/api/v1/auth/google/callback

JWT_SECRET=secret

FRONTEND_URL=http://localhost:5173
```

### 3. Setup Database

```bash
psql -U postgres -c "CREATE DATABASE listmak_service;"
```

### 4. Install Dependencies

```bash
go mod download
```

### 5. Run Application

```bash
# Development dengan hot-reload (gunakan Air)
air

# Atau langsung run
go run ./cmd/api
```

Server akan berjalan di `http://localhost:9001`

## API Documentation

Swagger documentation tersedia di:

```
http://localhost:9001/swagger/index.html
```

---

# Frontend

Vue 3 SPA dibangun dengan Vite.

## Quick Start

```bash
cd frontend
npm install
npm run dev
```

Dev server berjalan di `http://localhost:5173`

Lihat [Vue 3 script setup docs](https://v3.vuejs.org/api/sfc-script-setup.html#sfc-script-setup) untuk informasi lebih lanjut.

---

# Panduan Deployment ke Ubuntu Server

Panduan lengkap untuk menjalankan Listmak Service di server Ubuntu dengan Nginx sebagai reverse proxy dan PostgreSQL sebagai database.

## Prasyarat

- Server Ubuntu 20.04/22.04 LTS
- Akses SSH ke server dengan hak sudo
- Domain atau IP publik server

---

## 1. Update Sistem dan Install Dependensi Dasar

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget git build-essential
```

---

## 2. Install Golang

```bash
wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz

echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

go version
```

---

## 3. Install PostgreSQL Server

```bash
sudo apt install -y postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

### Buat Database dan User

```bash
sudo -u postgres psql
```

```sql
CREATE DATABASE listmak_service;
CREATE USER listmak_user WITH ENCRYPTED PASSWORD 'password_kuat_anda';
GRANT ALL PRIVILEGES ON DATABASE listmak_service TO listmak_user;
\q
```

---

## 4. Setup Project Golang

### Clone Repository

```bash
sudo mkdir -p /var/www/listmak-service
sudo chown $USER:$USER /var/www/listmak-service

cd /var/www
git clone <URL_REPOSITORY> listmak-service
cd listmak-service
```

### Konfigurasi Environment File

```bash
cd /var/www/listmak-service
cp .env.example .env
nano .env
```

```env
PORT=9001
ENV=production
DEBUG=false

DB_HOST=localhost
DB_NAME=listmak_service
DB_PORT=5432
DB_USER=listmak_user
DB_PASS=password_kuat_anda

GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=https://yourdomain.com/api/v1/auth/google/callback

JWT_SECRET=your_super_secret_jwt_key_yang_panjang_dan_aman

FRONTEND_URL=https://yourdomain.com
```

### Build Aplikasi

```bash
go mod download
go build -o listmak-service ./cmd/api
./listmak-service
```

---

## 5. Setup Systemd Service

```bash
sudo nano /etc/systemd/system/listmak-service.service
```

```ini
[Unit]
Description=Listmak Service Golang API
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/var/www/listmak-service
ExecStart=/var/www/listmak-service/listmak-service
Restart=always
RestartSec=5
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=listmak-service

Environment=PORT=9001
Environment=ENV=production

[Install]
WantedBy=multi-user.target
```

```bash
sudo chown -R www-data:www-data /var/www/listmak-service
sudo systemctl daemon-reload
sudo systemctl enable listmak-service
sudo systemctl start listmak-service
sudo systemctl status listmak-service
```

---

## 6. Install dan Konfigurasi Nginx

```bash
sudo apt install -y nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```

```bash
sudo nano /etc/nginx/sites-available/listmak-service
```

```nginx
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;

    access_log /var/log/nginx/listmak-service-access.log;
    error_log /var/log/nginx/listmak-service-error.log;

    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;

    location / {
        proxy_pass http://127.0.0.1:9001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;

        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    client_max_body_size 10M;
}
```

```bash
sudo ln -s /etc/nginx/sites-available/listmak-service /etc/nginx/sites-enabled/
sudo rm /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl reload nginx
```

---

## 7. Setup SSL dengan Let's Encrypt

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com
sudo systemctl status certbot.timer
sudo certbot renew --dry-run
```

---

## 8. Konfigurasi Firewall (UFW)

```bash
sudo ufw enable
sudo ufw allow OpenSSH
sudo ufw allow 'Nginx Full'
sudo ufw status
```

---

## 9. Commands Berguna untuk Maintenance

### Manage Service

```bash
sudo systemctl start listmak-service
sudo systemctl stop listmak-service
sudo systemctl restart listmak-service
sudo systemctl status listmak-service
```

### Lihat Logs

```bash
sudo journalctl -u listmak-service -f
sudo tail -f /var/log/nginx/listmak-service-access.log
sudo tail -f /var/log/nginx/listmak-service-error.log
sudo tail -f /var/log/postgresql/postgresql-*.log
```

### Update Aplikasi

```bash
cd /var/www/listmak-service
sudo systemctl stop listmak-service
git pull origin main
go build -o listmak-service ./cmd/api
sudo systemctl start listmak-service
```

---

## 10. Struktur Direktori Final

```
/var/www/listmak-service/
├── .env
├── listmak-service
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── configs/
│   ├── handlers/
│   ├── models/
│   └── routes/
├── go.mod
└── go.sum
```

---

## Troubleshooting

### Aplikasi tidak bisa connect ke PostgreSQL

```bash
sudo systemctl status postgresql
psql -U listmak_user -h localhost -d listmak_service
```

### Port sudah digunakan

```bash
sudo lsof -i :9001
sudo kill -9 <PID>
```

### Permission denied pada aplikasi

```bash
sudo chown -R www-data:www-data /var/www/listmak-service
sudo chmod +x /var/www/listmak-service/listmak-service
```

### Nginx 502 Bad Gateway

```bash
sudo systemctl status listmak-service
sudo netstat -tlpn | grep 9001
sudo systemctl restart listmak-service
sudo systemctl reload nginx
```

---

## Checklist Deployment

- [ ] Server Ubuntu sudah ready
- [ ] Golang terinstall
- [ ] PostgreSQL terinstall dan database dibuat
- [ ] Project sudah di-clone/upload
- [ ] File `.env` sudah dikonfigurasi
- [ ] Binary sudah di-build
- [ ] Systemd service sudah aktif
- [ ] Nginx sudah dikonfigurasi
- [ ] SSL sudah terpasang (optional)
- [ ] Firewall sudah dikonfigurasi
- [ ] Aplikasi bisa diakses dari browser

---

## License

MIT License

## Author

Muhammad Ali
