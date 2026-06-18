# Listmak Service API

REST API Server untuk manajemen Listmak, dibangun dengan Go (Gin Framework), MySQL, dan GORM.

## Tech Stack

- **Go** - Programming language
- **Gin** - Web framework
- **GORM** - ORM library
- **MySQL** - Primary database
- **JWT** - Authentication
- **Swagger** - API documentation

## Requirements

- Go 1.21+
- MySQL 8.0+
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
DB_PORT=3306
DB_USER=root
DB_PASS=your_password

GOOGLE_CLIENT_ID=your_id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your_secret
GOOGLE_REDIRECT_URL=http://localhost:9001/api/v1/auth/google/callback

JWT_SECRET=secret
```

### 3. Setup Database

```sql
CREATE DATABASE listmak_service CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
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

# Panduan Deployment ke Ubuntu Server

Panduan lengkap untuk menjalankan Listmak Service di server Ubuntu dengan Nginx sebagai reverse proxy dan MySQL sebagai database.

## Prasyarat

- Server Ubuntu 20.04/22.04 LTS
- Akses SSH ke server dengan hak sudo
- Domain atau IP publik server

---

## 1. Update Sistem dan Install Dependensi Dasar

```bash
# Update package list
sudo apt update && sudo apt upgrade -y

# Install dependensi dasar
sudo apt install -y curl wget git build-essential
```

---

## 2. Install Golang

```bash
# Download dan install Go (sesuaikan versi jika perlu)
wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz

# Extract ke /usr/local
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz

# Setup environment variables
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# Verifikasi instalasi
go version
```

---

## 3. Install MySQL Server

```bash
# Install MySQL
sudo apt install -y mysql-server

# Amankan instalasi MySQL
sudo mysql_secure_installation
```

Saat menjalankan `mysql_secure_installation`:

- Set root password
- Remove anonymous users: **Yes**
- Disallow root login remotely: **Yes** (untuk keamanan)
- Remove test database: **Yes**
- Reload privilege tables: **Yes**

### Buat Database dan User

```bash
# Login ke MySQL
sudo mysql -u root -p
```

Jalankan perintah SQL berikut:

```sql
-- Buat database
CREATE DATABASE listmak_service CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Buat user khusus untuk aplikasi
CREATE USER 'listmak_user'@'localhost' IDENTIFIED BY 'password_kuat_anda';

-- Berikan hak akses
GRANT ALL PRIVILEGES ON listmak_service.* TO 'listmak_user'@'localhost';

-- Flush privileges
FLUSH PRIVILEGES;

-- Exit
EXIT;
```

---

## 4. Setup Project Golang

### Clone Repository

```bash
# Buat direktori untuk aplikasi
sudo mkdir -p /var/www/listmak-service
sudo chown $USER:$USER /var/www/listmak-service

# Clone project (ganti dengan URL repo Anda)
cd /var/www
git clone <URL_REPOSITORY> listmak-service
cd listmak-service
```

Atau jika upload manual:

```bash
# Upload menggunakan SCP dari local machine
scp -r /path/to/listmak-service user@server_ip:/var/www/
```

### Konfigurasi Environment File

```bash
# Buat file .env
cd /var/www/listmak-service
cp .env.example .env
nano .env
```

Sesuaikan isi `.env` untuk production:

```env
PORT=9001
ENV=production
DEBUG=false

DB_HOST=localhost
DB_NAME=listmak_service
DB_PORT=3306
DB_USER=listmak_user
DB_PASS=password_kuat_anda

GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=https://yourdomain.com/api/v1/auth/google/callback

JWT_SECRET=your_super_secret_jwt_key_yang_panjang_dan_aman
```

### Build Aplikasi

```bash
# Download dependencies
go mod download

# Build binary
go build -o listmak-service ./cmd/api

# Verifikasi binary
./listmak-service
# Tekan Ctrl+C untuk berhenti
```

---

## 5. Setup Systemd Service

Buat service file agar aplikasi berjalan otomatis dan dapat di-manage dengan mudah.

```bash
sudo nano /etc/systemd/system/listmak-service.service
```

Isi dengan konfigurasi berikut:

```ini
[Unit]
Description=Listmak Service Golang API
After=network.target mysql.service
Wants=mysql.service

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

# Environment variables (opsional, bisa juga dari .env)
Environment=PORT=9001
Environment=ENV=production

[Install]
WantedBy=multi-user.target
```

### Aktifkan dan Jalankan Service

```bash
# Set permission yang tepat
sudo chown -R www-data:www-data /var/www/listmak-service

# Reload systemd daemon
sudo systemctl daemon-reload

# Enable service untuk start saat boot
sudo systemctl enable listmak-service

# Start service
sudo systemctl start listmak-service

# Cek status
sudo systemctl status listmak-service

# Lihat logs jika ada masalah
sudo journalctl -u listmak-service -f
```

---

## 6. Install dan Konfigurasi Nginx

### Install Nginx

```bash
sudo apt install -y nginx

# Start dan enable Nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```

### Konfigurasi Nginx sebagai Reverse Proxy

```bash
sudo nano /etc/nginx/sites-available/listmak-service
```

Isi dengan konfigurasi berikut:

```nginx
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;  # Ganti dengan domain/IP Anda

    # Logging
    access_log /var/log/nginx/listmak-service-access.log;
    error_log /var/log/nginx/listmak-service-error.log;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;

    # Proxy settings
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

        # Timeout settings
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Handle large request bodies
    client_max_body_size 10M;
}
```

### Aktifkan Konfigurasi

```bash
# Buat symbolic link ke sites-enabled
sudo ln -s /etc/nginx/sites-available/listmak-service /etc/nginx/sites-enabled/

# Hapus default config (opsional)
sudo rm /etc/nginx/sites-enabled/default

# Test konfigurasi Nginx
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

---

## 7. Setup SSL dengan Let's Encrypt (Opsional tapi Sangat Disarankan)

```bash
# Install Certbot
sudo apt install -y certbot python3-certbot-nginx

# Generate SSL certificate
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com

# Certbot akan otomatis mengupdate konfigurasi Nginx
# Auto-renewal sudah terinstall, tapi bisa dicek:
sudo systemctl status certbot.timer

# Test renewal
sudo certbot renew --dry-run
```

---

## 8. Konfigurasi Firewall (UFW)

```bash
# Enable UFW
sudo ufw enable

# Allow SSH
sudo ufw allow OpenSSH

# Allow HTTP dan HTTPS
sudo ufw allow 'Nginx Full'

# Atau allow port spesifik
# sudo ufw allow 80/tcp
# sudo ufw allow 443/tcp

# Verifikasi status
sudo ufw status
```

---

## 9. Commands Berguna untuk Maintenance

### Manage Service

```bash
# Start/Stop/Restart service
sudo systemctl start listmak-service
sudo systemctl stop listmak-service
sudo systemctl restart listmak-service

# Reload tanpa downtime (jika supported)
sudo systemctl reload listmak-service

# Cek status
sudo systemctl status listmak-service
```

### Lihat Logs

```bash
# Logs aplikasi
sudo journalctl -u listmak-service -f

# Logs Nginx
sudo tail -f /var/log/nginx/listmak-service-access.log
sudo tail -f /var/log/nginx/listmak-service-error.log

# Logs MySQL
sudo tail -f /var/log/mysql/error.log
```

### Update Aplikasi

```bash
cd /var/www/listmak-service

# Stop service
sudo systemctl stop listmak-service

# Pull latest code
git pull origin main

# Rebuild
go build -o listmak-service ./cmd/api

# Start service
sudo systemctl start listmak-service
```

---

## 10. Struktur Direktori Final

```
/var/www/listmak-service/
├── .env                    # Environment configuration
├── listmak-service         # Compiled binary
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

### Aplikasi tidak bisa connect ke MySQL

```bash
# Pastikan MySQL running
sudo systemctl status mysql

# Test koneksi manual
mysql -u listmak_user -p -h localhost listmak_service
```

### Port sudah digunakan

```bash
# Cek proses yang menggunakan port
sudo lsof -i :9001

# Kill proses jika perlu
sudo kill -9 <PID>
```

### Permission denied pada aplikasi

```bash
# Perbaiki ownership
sudo chown -R www-data:www-data /var/www/listmak-service

# Pastikan binary executable
sudo chmod +x /var/www/listmak-service/listmak-service
```

### Nginx 502 Bad Gateway

```bash
# Pastikan aplikasi berjalan
sudo systemctl status listmak-service

# Cek apakah listening di port yang benar
sudo netstat -tlpn | grep 9001

# Restart keduanya
sudo systemctl restart listmak-service
sudo systemctl reload nginx
```

---

## Checklist Deployment ✅

- [ ] Server Ubuntu sudah ready
- [ ] Golang terinstall
- [ ] MySQL terinstall dan database dibuat
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
