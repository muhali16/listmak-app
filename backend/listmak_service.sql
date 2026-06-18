-- MySQL dump 10.13  Distrib 8.4.6, for Linux (x86_64)
--
-- Host: localhost    Database: listmak_service
-- ------------------------------------------------------
-- Server version	8.4.6

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `contacts`
--

DROP TABLE IF EXISTS `contacts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `contacts` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `email` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `notes` text COLLATE utf8mb4_unicode_ci,
  `is_favorite` tinyint(1) DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`),
  KEY `idx_contacts_user_id` (`user_id`),
  KEY `idx_contacts_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `contacts`
--

LOCK TABLES `contacts` WRITE;
/*!40000 ALTER TABLE `contacts` DISABLE KEYS */;
/*!40000 ALTER TABLE `contacts` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `daily_summaries`
--

DROP TABLE IF EXISTS `daily_summaries`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `daily_summaries` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `date` date NOT NULL,
  `total_listmaks` int DEFAULT '0',
  `total_orders` int DEFAULT '0',
  `total_amount` decimal(12,2) DEFAULT '0.00',
  `paid_amount` decimal(12,2) DEFAULT '0.00',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`),
  UNIQUE KEY `date` (`date`),
  UNIQUE KEY `idx_daily_summaries_date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `daily_summaries`
--

LOCK TABLES `daily_summaries` WRITE;
/*!40000 ALTER TABLE `daily_summaries` DISABLE KEYS */;
/*!40000 ALTER TABLE `daily_summaries` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `listmaks`
--

DROP TABLE IF EXISTS `listmaks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `listmaks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `date` date NOT NULL,
  `created_by` bigint unsigned DEFAULT NULL,
  `total_orders` int DEFAULT '0',
  `total_amount` decimal(12,2) DEFAULT '0.00',
  `paid_amount` decimal(12,2) DEFAULT '0.00',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'active',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`),
  KEY `idx_listmaks_date` (`date`),
  KEY `idx_listmaks_created_by` (`created_by`),
  KEY `idx_listmaks_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_listmaks_user` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `listmaks`
--

LOCK TABLES `listmaks` WRITE;
/*!40000 ALTER TABLE `listmaks` DISABLE KEYS */;
INSERT INTO `listmaks` VALUES (1,'ListMak Minggu, 4 Januari','2026-01-04',5,4,68000.00,68000.00,'active','2026-01-04 22:01:49.259','2026-01-04 23:43:59.953',NULL),(3,'Listmak sore','2026-01-04',5,11,0.00,0.00,'active','2026-01-04 22:56:39.047','2026-01-04 23:47:45.597',NULL),(4,'listjum','2026-01-02',5,0,0.00,0.00,'active','2026-01-04 23:40:31.124','2026-01-04 23:40:31.124',NULL),(5,'Makan siang 31','2025-12-30',5,9,85000.00,0.00,'active','2026-01-04 23:52:29.350','2026-01-04 23:54:20.760',NULL);
/*!40000 ALTER TABLE `listmaks` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `order_history`
--

DROP TABLE IF EXISTS `order_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `order_history` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_id` int NOT NULL,
  `action` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `old_data` json DEFAULT NULL,
  `new_data` json DEFAULT NULL,
  `changed_by` int DEFAULT NULL,
  `changed_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`),
  KEY `idx_order_history_order_id` (`order_id`),
  KEY `idx_order_history_changed_at` (`changed_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `order_history`
--

LOCK TABLES `order_history` WRITE;
/*!40000 ALTER TABLE `order_history` DISABLE KEYS */;
/*!40000 ALTER TABLE `order_history` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `orders`
--

DROP TABLE IF EXISTS `orders`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `orders` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `listmak_id` bigint unsigned NOT NULL,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `order_detail` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `price` decimal(12,2) DEFAULT '0.00',
  `qty` int DEFAULT '1',
  `total_price` decimal(12,2) GENERATED ALWAYS AS ((`price` * `qty`)) STORED,
  `is_paid` tinyint(1) DEFAULT '0',
  `paid_at` datetime(3) DEFAULT NULL,
  `added_via` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'parse',
  `added_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`),
  KEY `idx_orders_listmak_id` (`listmak_id`),
  KEY `idx_orders_is_paid` (`is_paid`),
  KEY `idx_orders_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=33 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `orders`
--

LOCK TABLES `orders` WRITE;
/*!40000 ALTER TABLE `orders` DISABLE KEYS */;
INSERT INTO `orders` (`id`, `listmak_id`, `name`, `order_detail`, `price`, `qty`, `is_paid`, `paid_at`, `added_via`, `added_at`, `updated_at`, `deleted_at`) VALUES (8,1,'Mba Ribka','Nasi ayam madura',22000.00,1,1,'2026-01-04 23:43:59.665','','2026-01-04 15:55:18','2026-01-04 16:44:00',NULL),(9,1,'Dimas','sop ayam pak min + nasi',12000.00,1,0,NULL,'','2026-01-04 16:01:01','2026-01-04 16:07:44','2026-01-04 16:15:48'),(10,1,'Ali','nasi ayam kremez sambel hijau + nasi dua bungkus',23000.00,1,1,'2026-01-04 23:43:59.867','','2026-01-04 16:01:01','2026-01-04 16:44:00',NULL),(11,1,'dhani','cimol bojot ',23000.00,1,1,'2026-01-04 23:43:59.786','','2026-01-04 16:19:21','2026-01-04 16:44:00',NULL),(12,1,'Ichaa','Nasi ayam madura',0.00,1,0,NULL,'parse','2026-01-04 16:29:33','2026-01-04 16:44:00',NULL),(13,3,'⁠Ali','nasi bu mul pake urap + tempe orek basah + 1 bakwan jagung',0.00,1,0,NULL,'parse','2026-01-04 16:30:44','2026-01-04 16:30:46',NULL),(14,3,'⁠Icha','Sop Ayam Pak Min PAHA (Gak pake nasi)',0.00,1,0,NULL,'','2026-01-04 16:30:44','2026-01-04 16:30:47',NULL),(15,3,'Safira','nasi ayam Madura paha',0.00,1,0,NULL,'parse','2026-01-04 16:30:44','2026-01-04 16:30:46',NULL),(16,3,'Nadiyah','nasi bu mul 1/2 + urap',0.00,1,0,NULL,'parse','2026-01-04 16:30:44','2026-01-04 16:30:46',NULL),(17,3,'rachel','sop ayam pak min PAHA + nasi',0.00,1,0,NULL,'parse','2026-01-04 16:30:44','2026-01-04 16:30:46',NULL),(18,3,'Jo','Sop Ayam Pak Min PAHA + nasi + tempe',0.00,1,0,NULL,'parse','2026-01-04 16:30:44','2026-01-04 16:30:46',NULL),(19,3,'⁠Reni','nasi ayam Madura dada tidak pakai sambel',0.00,1,0,NULL,'','2026-01-04 16:30:45','2026-01-04 16:30:47',NULL),(20,3,'Susan','nasi bu mul + pepes ikan mas',0.00,1,1,'2026-01-04 23:47:45.574','','2026-01-04 16:30:45','2026-01-04 16:47:46',NULL),(21,3,'Mona','nasi + kentang Mustofa (5rb) dipisah',0.00,1,1,'2026-01-04 23:47:44.353','','2026-01-04 16:30:45','2026-01-04 16:47:44',NULL),(22,3,'Icha','nasi goreng',0.00,1,1,'2026-01-04 23:47:43.374','','2026-01-04 16:46:58','2026-01-04 16:47:43',NULL),(23,3,'Dhani','ayam goreng',0.00,1,1,'2026-01-04 23:47:41.914','','2026-01-04 16:46:58','2026-01-04 16:47:42',NULL),(24,5,'⁠Ali','nasi bu mul pake urap + tempe orek basah + 1 bakwan jagung',20000.00,1,0,NULL,'parse','2026-01-04 16:53:43','2026-01-04 16:54:03',NULL),(25,5,'⁠Icha','Sop Ayam Pak Min PAHA (Gak pake nasi)',40000.00,1,0,NULL,'parse','2026-01-04 16:53:43','2026-01-04 16:54:13',NULL),(26,5,'Safira','nasi ayam Madura paha',25000.00,1,0,NULL,'parse','2026-01-04 16:53:43','2026-01-04 16:54:21',NULL),(27,5,'Nadiyah','nasi bu mul 1/2 + urap',0.00,1,0,NULL,'manual','2026-01-04 16:53:43','2026-01-04 16:53:43',NULL),(28,5,'rachel','sop ayam pak min PAHA + nasi',0.00,1,0,NULL,'manual','2026-01-04 16:53:43','2026-01-04 16:53:43',NULL),(29,5,'Jo','Sop Ayam Pak Min PAHA + nasi + tempe',0.00,1,0,NULL,'manual','2026-01-04 16:53:43','2026-01-04 16:53:43',NULL),(30,5,'⁠Reni','nasi ayam Madura dada tidak pakai sambel',0.00,1,0,NULL,'manual','2026-01-04 16:53:43','2026-01-04 16:53:43',NULL),(31,5,'Susan','nasi bu mul + pepes ikan mas',0.00,1,0,NULL,'manual','2026-01-04 16:53:43','2026-01-04 16:53:43',NULL),(32,5,'Mona','nasi + kentang Mustofa (5rb) dipisah',0.00,1,0,NULL,'manual','2026-01-04 16:53:43','2026-01-04 16:53:43',NULL);
/*!40000 ALTER TABLE `orders` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `share_links`
--

DROP TABLE IF EXISTS `share_links`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `share_links` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `share_id` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `listmak_id` int NOT NULL,
  `title` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `expires_at` timestamp NOT NULL,
  `is_active` tinyint(1) DEFAULT '1',
  `created_by` int DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`),
  UNIQUE KEY `share_id` (`share_id`),
  UNIQUE KEY `idx_share_links_share_id` (`share_id`),
  KEY `idx_share_links_listmak_id` (`listmak_id`),
  KEY `idx_share_links_expires_at` (`expires_at`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `share_links`
--

LOCK TABLES `share_links` WRITE;
/*!40000 ALTER TABLE `share_links` DISABLE KEYS */;
INSERT INTO `share_links` VALUES (3,'awF8feYu',1,'Makan siang hari ini','2026-01-04 15:42:00',1,5,'2026-01-04 15:38:17','2026-01-04 15:38:16'),(4,'8ztm8aV3',1,'Input ListMak 4/1/2026','2026-01-04 16:20:00',1,5,'2026-01-04 16:19:03','2026-01-04 16:19:03'),(5,'0KCCq6bR',1,'Input ListMak 4/1/2026','2026-01-04 05:24:00',1,5,'2026-01-04 16:24:37','2026-01-04 16:24:37'),(6,'Z00cecbN',1,'Input ListMak 4/1/2026','2026-01-04 05:27:00',1,5,'2026-01-04 16:27:25','2026-01-04 16:27:24'),(7,'tuV6c6L6',1,'Input ListMak 4/1/2026','2026-01-03 17:27:00',1,5,'2026-01-04 16:27:51','2026-01-04 16:27:51'),(8,'Sxx9bjnR',1,'Input ListMak 4/1/2026','2026-01-04 17:28:00',1,5,'2026-01-04 16:28:20','2026-01-04 16:28:19'),(9,'NjoEtbwC',3,'Input ListMak 4/1/2026','2026-01-04 16:50:00',1,5,'2026-01-04 16:45:11','2026-01-04 16:45:10');
/*!40000 ALTER TABLE `share_links` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `system_logs`
--

DROP TABLE IF EXISTS `system_logs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `request_id` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `method` longtext COLLATE utf8mb4_unicode_ci,
  `path` longtext COLLATE utf8mb4_unicode_ci,
  `status_code` bigint DEFAULT NULL,
  `latency` longtext COLLATE utf8mb4_unicode_ci,
  `client_ip` longtext COLLATE utf8mb4_unicode_ci,
  `error_msg` longtext COLLATE utf8mb4_unicode_ci,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_system_logs_request_id` (`request_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `system_logs`
--

LOCK TABLES `system_logs` WRITE;
/*!40000 ALTER TABLE `system_logs` DISABLE KEYS */;
/*!40000 ALTER TABLE `system_logs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `google_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `avatar` text COLLATE utf8mb4_unicode_ci,
  `role` enum('admin','user') COLLATE utf8mb4_unicode_ci DEFAULT 'user',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_email` (`email`),
  UNIQUE KEY `idx_users_google_id` (`google_id`),
  KEY `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (5,'101414767172550559118','muhammadali55214@gmail.com','Muhammad Ali Mustaqim','https://lh3.googleusercontent.com/a/ACg8ocK_8bjXy7iyjYW9T6KyACCqVSCSkr3PvURGSvrWspxaF_3F-zPH=s96-c','user','2025-12-27 17:04:07.635','2025-12-27 17:04:07.635',NULL),(6,'112764661582594114435','muhammadali55214.mri@gmail.com','Muhammad Ali Mustaqim MRI','https://lh3.googleusercontent.com/a/ACg8ocIUDlMKUur_1Hwhcxg3E_2OIu6mQVgSkL3WgzEjW3Bys1ba=s96-c','user','2026-01-04 19:51:34.354','2026-01-04 19:51:34.354',NULL);
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `view_shares`
--

DROP TABLE IF EXISTS `view_shares`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `view_shares` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `view_id` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `listmak_id` int NOT NULL,
  `title` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `snapshot_data` json DEFAULT NULL,
  `created_by` int DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`),
  UNIQUE KEY `view_id` (`view_id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `view_shares`
--

LOCK TABLES `view_shares` WRITE;
/*!40000 ALTER TABLE `view_shares` DISABLE KEYS */;
INSERT INTO `view_shares` VALUES (2,'yKSVvLx1',1,'ListMak Minggu, 4 Januari','{\"id\": 1, \"date\": \"2026-01-04T00:00:00+07:00\", \"user\": {\"id\": 5, \"name\": \"Muhammad Ali Mustaqim\", \"role\": \"user\", \"email\": \"muhammadali55214@gmail.com\", \"avatar\": \"https://lh3.googleusercontent.com/a/ACg8ocK_8bjXy7iyjYW9T6KyACCqVSCSkr3PvURGSvrWspxaF_3F-zPH=s96-c\", \"google_id\": \"101414767172550559118\", \"created_at\": \"2025-12-27T17:04:07.635+07:00\", \"updated_at\": \"2025-12-27T17:04:07.635+07:00\"}, \"title\": \"ListMak Minggu, 4 Januari\", \"orders\": [{\"id\": 8, \"qty\": 1, \"name\": \"Mba Ribka\", \"price\": 22000, \"is_paid\": false, \"paid_at\": null, \"added_at\": \"2026-01-04T22:55:18+07:00\", \"added_via\": \"manual\", \"listmak_id\": 1, \"updated_at\": \"2026-01-04T22:55:19+07:00\", \"total_price\": 22000, \"order_detail\": \"Nasi ayam madura\"}], \"status\": \"active\", \"created_at\": \"2026-01-04T22:01:49.259+07:00\", \"created_by\": 5, \"updated_at\": \"2026-01-04T22:55:18.799+07:00\", \"paid_amount\": 0, \"total_amount\": 22000, \"total_orders\": 1}',5,'2026-01-04 15:59:56'),(3,'e0kbJ2Gp',1,'ListMak Minggu, 4 Januari','{\"id\": 1, \"date\": \"2026-01-04T00:00:00+07:00\", \"user\": {\"id\": 5, \"name\": \"Muhammad Ali Mustaqim\", \"role\": \"user\", \"email\": \"muhammadali55214@gmail.com\", \"avatar\": \"https://lh3.googleusercontent.com/a/ACg8ocK_8bjXy7iyjYW9T6KyACCqVSCSkr3PvURGSvrWspxaF_3F-zPH=s96-c\", \"google_id\": \"101414767172550559118\", \"created_at\": \"2025-12-27T17:04:07.635+07:00\", \"updated_at\": \"2025-12-27T17:04:07.635+07:00\"}, \"title\": \"ListMak Minggu, 4 Januari\", \"orders\": [{\"id\": 8, \"qty\": 1, \"name\": \"Mba Ribka\", \"price\": 22000, \"is_paid\": false, \"paid_at\": null, \"added_at\": \"2026-01-04T22:55:18+07:00\", \"added_via\": \"parse\", \"listmak_id\": 1, \"updated_at\": \"2026-01-04T23:01:01+07:00\", \"total_price\": 22000, \"order_detail\": \"Nasi ayam madura\"}, {\"id\": 9, \"qty\": 1, \"name\": \"Dimas\", \"price\": 0, \"is_paid\": false, \"paid_at\": null, \"added_at\": \"2026-01-04T23:01:01+07:00\", \"added_via\": \"manual\", \"listmak_id\": 1, \"updated_at\": \"2026-01-04T23:01:01+07:00\", \"total_price\": 0, \"order_detail\": \"sop ayam pak min + nasi\"}, {\"id\": 10, \"qty\": 1, \"name\": \"Ali\", \"price\": 0, \"is_paid\": false, \"paid_at\": null, \"added_at\": \"2026-01-04T23:01:01+07:00\", \"added_via\": \"manual\", \"listmak_id\": 1, \"updated_at\": \"2026-01-04T23:01:01+07:00\", \"total_price\": 0, \"order_detail\": \"nasi ayam kremez sambel hijau + nasi\"}], \"status\": \"active\", \"created_at\": \"2026-01-04T22:01:49.259+07:00\", \"created_by\": 5, \"updated_at\": \"2026-01-04T23:01:01.199+07:00\", \"paid_amount\": 0, \"total_amount\": 0, \"total_orders\": 2}',5,'2026-01-04 16:02:01'),(4,'IaJ0o59o',1,'ListMak Minggu, 4 Januari','{\"id\": 1, \"date\": \"2026-01-04T00:00:00+07:00\", \"user\": {\"id\": 5, \"name\": \"Muhammad Ali Mustaqim\", \"role\": \"user\", \"email\": \"muhammadali55214@gmail.com\", \"avatar\": \"https://lh3.googleusercontent.com/a/ACg8ocK_8bjXy7iyjYW9T6KyACCqVSCSkr3PvURGSvrWspxaF_3F-zPH=s96-c\", \"google_id\": \"101414767172550559118\", \"created_at\": \"2025-12-27T17:04:07.635+07:00\", \"updated_at\": \"2025-12-27T17:04:07.635+07:00\"}, \"title\": \"ListMak Minggu, 4 Januari\", \"orders\": [{\"id\": 8, \"qty\": 1, \"name\": \"Mba Ribka\", \"price\": 22000, \"is_paid\": false, \"paid_at\": null, \"added_at\": \"2026-01-04T22:55:18+07:00\", \"added_via\": \"\", \"listmak_id\": 1, \"updated_at\": \"2026-01-04T23:12:28+07:00\", \"total_price\": 22000, \"order_detail\": \"Nasi ayam madura\"}, {\"id\": 10, \"qty\": 1, \"name\": \"Ali\", \"price\": 23000, \"is_paid\": true, \"paid_at\": \"2026-01-04T23:12:27.643+07:00\", \"added_at\": \"2026-01-04T23:01:01+07:00\", \"added_via\": \"\", \"listmak_id\": 1, \"updated_at\": \"2026-01-04T23:12:28+07:00\", \"total_price\": 23000, \"order_detail\": \"nasi ayam kremez sambel hijau + nasi dua bungkus\"}], \"status\": \"active\", \"created_at\": \"2026-01-04T22:01:49.259+07:00\", \"created_by\": 5, \"updated_at\": \"2026-01-04T23:15:48.49+07:00\", \"paid_amount\": 23000, \"total_amount\": 45000, \"total_orders\": 2}',5,'2026-01-04 16:16:01'),(5,'tquGoiSZ',1,'ListMak Minggu, 4 Januari','{\"id\": 1, \"date\": \"2026-01-04T00:00:00+07:00\", \"user\": {\"id\": 5, \"name\": \"Muhammad Ali Mustaqim\", \"role\": \"user\", \"email\": \"muhammadali55214@gmail.com\", \"avatar\": \"https://lh3.googleusercontent.com/a/ACg8ocK_8bjXy7iyjYW9T6KyACCqVSCSkr3PvURGSvrWspxaF_3F-zPH=s96-c\", \"google_id\": \"101414767172550559118\", \"created_at\": \"2025-12-27T17:04:07.635+07:00\", \"updated_at\": \"2025-12-27T17:04:07.635+07:00\"}, \"title\": \"ListMak Minggu, 4 Januari\", \"orders\": [{\"id\": 8, \"qty\": 1, \"name\": \"Mba Ribka\", \"price\": 22000, \"is_paid\": true, \"paid_at\": \"2026-01-04T23:43:59.665+07:00\", \"added_at\": \"2026-01-04T22:55:18+07:00\", \"added_via\": \"\", \"listmak_id\": 1, \"updated_at\": \"2026-01-04T23:44:00+07:00\", \"total_price\": 22000, \"order_detail\": \"Nasi ayam madura\"}, {\"id\": 10, \"qty\": 1, \"name\": \"Ali\", \"price\": 23000, \"is_paid\": true, \"paid_at\": \"2026-01-04T23:43:59.867+07:00\", \"added_at\": \"2026-01-04T23:01:01+07:00\", \"added_via\": \"\", \"listmak_id\": 1, \"updated_at\": \"2026-01-04T23:44:00+07:00\", \"total_price\": 23000, \"order_detail\": \"nasi ayam kremez sambel hijau + nasi dua bungkus\"}, {\"id\": 11, \"qty\": 1, \"name\": \"dhani\", \"price\": 23000, \"is_paid\": true, \"paid_at\": \"2026-01-04T23:43:59.786+07:00\", \"added_at\": \"2026-01-04T23:19:21+07:00\", \"added_via\": \"\", \"listmak_id\": 1, \"updated_at\": \"2026-01-04T23:44:00+07:00\", \"total_price\": 23000, \"order_detail\": \"cimol bojot \"}, {\"id\": 12, \"qty\": 1, \"name\": \"Ichaa\", \"price\": 0, \"is_paid\": false, \"paid_at\": null, \"added_at\": \"2026-01-04T23:29:33+07:00\", \"added_via\": \"parse\", \"listmak_id\": 1, \"updated_at\": \"2026-01-04T23:44:00+07:00\", \"total_price\": 0, \"order_detail\": \"Nasi ayam madura\"}], \"status\": \"active\", \"created_at\": \"2026-01-04T22:01:49.259+07:00\", \"created_by\": 5, \"updated_at\": \"2026-01-04T23:43:59.953+07:00\", \"paid_amount\": 68000, \"share_links\": [{\"id\": 8, \"title\": \"Input ListMak 4/1/2026\", \"listmak\": {\"id\": 0, \"date\": \"0001-01-01T00:00:00Z\", \"title\": \"\", \"status\": \"\", \"created_at\": \"0001-01-01T00:00:00Z\", \"created_by\": null, \"updated_at\": \"0001-01-01T00:00:00Z\", \"paid_amount\": 0, \"total_amount\": 0, \"total_orders\": 0}, \"share_id\": \"Sxx9bjnR\", \"is_active\": true, \"created_at\": \"2026-01-04T23:28:20+07:00\", \"created_by\": 5, \"expires_at\": \"2026-01-05T00:28:00+07:00\", \"listmak_id\": 1}], \"total_amount\": 68000, \"total_orders\": 4}',5,'2026-01-04 16:49:10');
/*!40000 ALTER TABLE `view_shares` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-01-16  1:24:40
