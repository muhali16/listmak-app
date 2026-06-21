package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/muhali16/listmak-service/internal/models"
	"github.com/muhali16/listmak-service/internal/repository"
)

type ParseOrderInput struct {
	Name        string `json:"name"`
	OrderDetail string `json:"order_detail"`
}

type ParsedOrderItem struct {
	Name           string `json:"name"`
	OrderDetail    string `json:"order_detail"`
	VendorName     string `json:"vendor_name"`
	EstimatedPrice int    `json:"estimated_price"`
	Qty            int    `json:"qty"`
}

type AIService interface {
	ExtractVendor(requestID string, orderDetail string, orderID *uint) (string, error)
	ExtractVendorsBatch(requestID string, orders []models.Order) (map[uint]string, error)
	SummarizeOrders(requestID string, orders []models.Order, catalog []models.PriceCatalog, existing *models.SummaryData, location string) (*models.SummaryData, error)
	EstimatePrice(requestID string, itemDetail string, location string) (int, error)
	ParseOrders(requestID string, orders []ParseOrderInput, location string) ([]ParsedOrderItem, error)
}

type fireworksAIService struct {
	apiKey  string
	model   string
	client  *http.Client
	logRepo repository.AILogRepository
}

func NewFireworksAIService(apiKey, model string, logRepo repository.AILogRepository) AIService {
	return &fireworksAIService{
		apiKey:  apiKey,
		model:   model,
		client:  &http.Client{Timeout: 10 * time.Second},
		logRepo: logRepo,
	}
}

type noopAIService struct{}

func NewNoopAIService() AIService { return &noopAIService{} }

func (n *noopAIService) ExtractVendor(_ string, _ string, _ *uint) (string, error) { return "", nil }
func (n *noopAIService) ExtractVendorsBatch(_ string, _ []models.Order) (map[uint]string, error) {
	return map[uint]string{}, nil
}
func (n *noopAIService) SummarizeOrders(_ string, _ []models.Order, _ []models.PriceCatalog, _ *models.SummaryData, _ string) (*models.SummaryData, error) {
	return &models.SummaryData{}, nil
}
func (n *noopAIService) EstimatePrice(_ string, _ string, _ string) (int, error) { return 0, nil }
func (n *noopAIService) ParseOrders(_ string, orders []ParseOrderInput, _ string) ([]ParsedOrderItem, error) {
	items := make([]ParsedOrderItem, len(orders))
	for i, o := range orders {
		items[i] = ParsedOrderItem{Name: o.Name, OrderDetail: o.OrderDetail}
	}
	return items, nil
}

func (s *fireworksAIService) ExtractVendor(requestID string, orderDetail string, orderID *uint) (string, error) {
	start := time.Now()

	prompt := fmt.Sprintf(
		"Tugas Anda adalah mengekstrak nama tempat atau kategori vendor dari pesanan makanan.\n\n"+
			"Aturan:\n"+
			"1. Jika ada nama orang/identitas unik yang digabung dengan jenis makanan (contoh: 'Bakso bakar pak donan', 'mie ayam bu nun'), ambil kombinasi makanan + nama tersebut menjadi 'Bakso Bakar Pak Donan' atau 'Mie Ayam Bu Nun'.\n"+
			"2. Jika berupa jajanan umum/kaki lima tanpa nama penjual spesifik (contoh: 'Cilok 5 tusuk', 'telur gulung'), cukup ambil nama jenis makanannya saja menjadi 'Cilok' atau 'Telur Gulung'.\n"+
			"3. Jika ada nama daerah/gaya masakan (contoh: 'ayam madura dada', 'nasi goreng madura'), ambil menjadi 'Ayam Madura' atau 'Nasi Goreng Madura'.\n\n"+
			"Format Output: Jawab HANYA nama tempat/vendor hasil ekstraksi tersebut, singkat, gunakan huruf kapital di awal kata (Title Case), tanpa tanda kutip, dan TANPA penjelasan apapun.\n\n"+
			"Pesanan: %s", orderDetail,
	)

	reqBody, _ := json.Marshal(map[string]any{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":       30,
		"temperature":      0,
		"reasoning_effort": "none",
	})

	req, err := http.NewRequest("POST", "https://api.fireworks.ai/inference/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		s.writeLog(requestID, orderID, orderDetail, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.writeLog(requestID, orderID, orderDetail, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("fireworks API %d: %s", resp.StatusCode, string(b))
		s.writeLog(requestID, orderID, orderDetail, "", time.Since(start).Milliseconds(), "failed", errMsg)
		return "", fmt.Errorf("%s", errMsg)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.writeLog(requestID, orderID, orderDetail, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return "", err
	}
	if len(result.Choices) == 0 {
		s.writeLog(requestID, orderID, orderDetail, "", time.Since(start).Milliseconds(), "success", "")
		return "", nil
	}

	output := strings.TrimSpace(result.Choices[0].Message.Content)
	s.writeLog(requestID, orderID, orderDetail, output, time.Since(start).Milliseconds(), "success", "")
	return output, nil
}

func (s *fireworksAIService) ExtractVendorsBatch(requestID string, orders []models.Order) (map[uint]string, error) {
	if len(orders) == 0 {
		return map[uint]string{}, nil
	}

	start := time.Now()

	details := make([]string, len(orders))
	for i, o := range orders {
		details[i] = o.OrderDetail
	}
	detailsJSON, _ := json.Marshal(details)
	input := string(detailsJSON)

	prompt := fmt.Sprintf(
		"Ekstrak nama vendor/tempat beli dari setiap pesanan makanan berikut.\n\n"+
			"Aturan:\n"+
			"1. Jika ada nama orang unik + jenis makanan (contoh: 'bakso bakar pak donan') → ambil kombinasinya: 'Bakso Bakar Pak Donan'\n"+
			"2. Jajanan umum tanpa nama penjual (contoh: 'cilok 5 tusuk') → nama jenis saja: 'Cilok'\n"+
			"3. Ada nama daerah/gaya masakan (contoh: 'ayam madura dada') → sertakan daerahnya: 'Ayam Madura'\n"+
			"4. Kata 'nasi' di awal pesanan adalah lauk pendamping, BUKAN bagian nama vendor — hilangkan, kecuali nama hidangannya memang 'nasi goreng', 'nasi uduk', 'nasi padang', 'nasi kuning', 'nasi bakar', atau sejenisnya. Contoh: 'nasi ayam madura dada' → 'Ayam Madura', bukan 'Nasi Ayam Madura'.\n"+
			"5. Tambahan topping/pelengkap (contoh: '+ kerupuk putih bulat', 'tidak pedas') diabaikan — fokus pada hidangan utama.\n"+
			"6. Pesanan berbeda yang merujuk tempat yang sama HARUS menghasilkan nama vendor yang sama persis. Contoh: 'ayam madura dada', 'nasi ayam madura dada', 'ayam madura paha' → semuanya 'Ayam Madura'.\n\n"+
			"Output: JSON array string SESUAI URUTAN input, Title Case, tanpa penjelasan, tanpa markdown.\n"+
			"Contoh input: [\"nasi ayam madura dada\",\"ayam madura paha\",\"cilok 5 tusuk\"]\n"+
			"Contoh output: [\"Ayam Madura\",\"Ayam Madura\",\"Cilok\"]\n\n"+
			"Input: %s", input,
	)

	reqBody, _ := json.Marshal(map[string]any{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":       1024,
		"temperature":      0,
		"reasoning_effort": "none",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.fireworks.ai/inference/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		s.writeLog(requestID, nil, input, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.writeLog(requestID, nil, input, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("fireworks API %d: %s", resp.StatusCode, string(b))
		s.writeLog(requestID, nil, input, "", time.Since(start).Milliseconds(), "failed", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.writeLog(requestID, nil, input, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return nil, err
	}
	if len(result.Choices) == 0 {
		s.writeLog(requestID, nil, input, "", time.Since(start).Milliseconds(), "success", "")
		return map[uint]string{}, nil
	}

	raw := strings.TrimSpace(result.Choices[0].Message.Content)
	// strip markdown fences if model wraps in ```json ... ```
	if idx := strings.Index(raw, "["); idx > 0 {
		raw = raw[idx:]
	}
	if idx := strings.LastIndex(raw, "]"); idx >= 0 && idx < len(raw)-1 {
		raw = raw[:idx+1]
	}

	var vendors []string
	if err := json.Unmarshal([]byte(raw), &vendors); err != nil {
		s.writeLog(requestID, nil, input, raw, time.Since(start).Milliseconds(), "failed", err.Error())
		return map[uint]string{}, nil
	}

	s.writeLog(requestID, nil, input, raw, time.Since(start).Milliseconds(), "success", "")

	out := make(map[uint]string, len(orders))
	for i, o := range orders {
		if i < len(vendors) {
			out[o.ID] = strings.TrimSpace(vendors[i])
		}
	}
	return out, nil
}

func (s *fireworksAIService) SummarizeOrders(requestID string, orders []models.Order, catalog []models.PriceCatalog, existing *models.SummaryData, location string) (*models.SummaryData, error) {
	if len(orders) == 0 {
		return existing, nil
	}

	start := time.Now()

	type orderInput struct {
		Name        string  `json:"name"`
		OrderDetail string  `json:"order_detail"`
		VendorName  string  `json:"vendor_name,omitempty"`
		Price       *int    `json:"price"` // null = belum ada harga
	}
	orderList := make([]orderInput, len(orders))
	for i, o := range orders {
		var price *int
		if o.Price > 0 {
			p := int(o.Price)
			price = &p
		}
		orderList[i] = orderInput{
			Name:        o.Name,
			OrderDetail: o.OrderDetail,
			VendorName:  o.VendorName,
			Price:       price,
		}
	}
	ordersJSON, _ := json.Marshal(orderList)
	catalogJSON, _ := json.Marshal(catalog)

	locationCtx := ""
	if location != "" {
		locationCtx = fmt.Sprintf("Konteks lokasi pengguna: %s — sesuaikan estimasi harga dengan standar harga warung/kaki lima di sekitar lokasi tersebut.\n\n", location)
	}

	var prompt string
	if existing != nil {
		existingJSON, _ := json.Marshal(existing)
		prompt = fmt.Sprintf(
			"Kamu adalah asisten ringkasan belanja kolektif.\n\n"+
				"%s"+
				"Ringkasan belanja yang sudah ada (JSON):\n%s\n\n"+
				"Order baru yang perlu ditambahkan (field 'price' = harga aktual dari pengguna, null = belum ada harga):\n%s\n\n"+
				"Katalog harga tersedia:\n%s\n\n"+
				"Perbarui ringkasan: merge item dari vendor yang sama, tambah qty jika item sudah ada, tambah item/vendor baru jika belum ada.\n"+
				"Untuk harga: jika order memiliki 'price' tidak null → gunakan sebagai unit_price, is_estimated: false. "+
				"Jika 'price' null → cek katalog (is_estimated: false) atau estimasi (is_estimated: true). "+
				"Pertahankan unit_price_actual yang sudah ada. Perbarui total_estimated.\n\n"+
				"Output JSON saja, tanpa penjelasan, tanpa markdown:\n"+
				`{"vendors":[{"name":"Nama Vendor","items":[{"name":"Nama Item","qty":1,"unit_price":15000,"is_estimated":true}]}],"total_estimated":75000}`,
			locationCtx, string(existingJSON), string(ordersJSON), string(catalogJSON),
		)
	} else {
		prompt = fmt.Sprintf(
			"Kamu adalah asisten ringkasan belanja kolektif.\n\n"+
				"%s"+
				"Buat ringkasan belanja terstruktur dari daftar order berikut.\n\n"+
				"Order list (field 'price' = harga aktual dari pengguna, null = belum ada harga):\n%s\n\n"+
				"Katalog harga tersedia (format: {\"vendor_name\":\"...\",\"item_name\":\"...\",\"price\":XXXX}):\n%s\n\n"+
				"Aturan:\n"+
				"1. Identifikasi vendor/lokasi dari konteks setiap order dan field vendor_name jika ada\n"+
				"2. Kelompokkan item per vendor yang sama\n"+
				"3. Normalisasi nama item: Title Case, tanpa nama pemesan, tanpa keterangan jumlah\n"+
				"4. 'nasi' sebagai lauk pendamping (bukan 'nasi goreng'/'nasi uduk'/'nasi padang'/dll) → item 'Nasi' di vendor yang sama\n"+
				"5. Agregasi item identik dari vendor yang sama → tambahkan qty\n"+
				"6. Topping/pelengkap (misal 'kerupuk', 'tahu', 'tempe') → item terpisah di vendor yang sama\n"+
				"7. Untuk unit_price: jika order memiliki 'price' tidak null → gunakan sebagai unit_price, set is_estimated: false\n"+
				"8. Jika 'price' null → cek katalog (is_estimated: false jika ada), estimasi jika tidak ada (is_estimated: true)\n"+
				"9. Jika ada 'price' dari pengguna tapi berbeda signifikan (>30%%) dari katalog → pertimbangkan gunakan harga katalog\n"+
				"10. Estimasi harga dalam Rupiah yang wajar untuk warung makan di sekitar lokasi\n"+
				"11. total_estimated = sum semua (unit_price * qty)\n\n"+
				"Output JSON saja, tanpa penjelasan, tanpa markdown:\n"+
				`{"vendors":[{"name":"Nama Vendor","items":[{"name":"Nama Item","qty":1,"unit_price":15000,"is_estimated":true}]}],"total_estimated":75000}`,
			locationCtx, string(ordersJSON), string(catalogJSON),
		)
	}

	reqBody, _ := json.Marshal(map[string]any{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":       2048,
		"temperature":      0,
		"reasoning_effort": "none",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.fireworks.ai/inference/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		s.writeLog(requestID, nil, prompt, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.writeLog(requestID, nil, prompt, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("fireworks API %d: %s", resp.StatusCode, string(b))
		s.writeLog(requestID, nil, prompt, "", time.Since(start).Milliseconds(), "failed", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.writeLog(requestID, nil, prompt, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return nil, err
	}
	if len(result.Choices) == 0 {
		s.writeLog(requestID, nil, string(ordersJSON), "", time.Since(start).Milliseconds(), "success", "")
		return &models.SummaryData{}, nil
	}

	raw := strings.TrimSpace(result.Choices[0].Message.Content)
	if idx := strings.Index(raw, "{"); idx > 0 {
		raw = raw[idx:]
	}
	if idx := strings.LastIndex(raw, "}"); idx >= 0 && idx < len(raw)-1 {
		raw = raw[:idx+1]
	}

	var summary models.SummaryData
	if err := json.Unmarshal([]byte(raw), &summary); err != nil {
		s.writeLog(requestID, nil, prompt, raw, time.Since(start).Milliseconds(), "failed", err.Error())
		return nil, err
	}

	s.writeLog(requestID, nil, prompt, raw, time.Since(start).Milliseconds(), "success", "")
	return &summary, nil
}

func (s *fireworksAIService) EstimatePrice(requestID string, itemDetail string, location string) (int, error) {
	start := time.Now()

	locationLine := ""
	if location != "" {
		locationLine = fmt.Sprintf("Lokasi pengguna: %s — sesuaikan harga dengan standar warung/kaki lima di sekitar lokasi tersebut.\n", location)
	}

	prompt := fmt.Sprintf(
		"Berikan estimasi harga yang wajar untuk item makanan/minuman berikut di warung makan atau kaki lima Indonesia.\n"+
			"%s"+
			"Pesanan: %s\n"+
			"Jawab HANYA angka integer (harga dalam Rupiah, tanpa titik/koma/Rp). Contoh: 15000",
		locationLine, itemDetail,
	)

	reqBody, _ := json.Marshal(map[string]any{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":       20,
		"temperature":      0,
		"reasoning_effort": "none",
	})

	req, err := http.NewRequest("POST", "https://api.fireworks.ai/inference/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		s.writeLog(requestID, nil, prompt, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.writeLog(requestID, nil, prompt, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("fireworks API %d: %s", resp.StatusCode, string(b))
		s.writeLog(requestID, nil, prompt, "", time.Since(start).Milliseconds(), "failed", errMsg)
		return 0, fmt.Errorf("%s", errMsg)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.writeLog(requestID, nil, prompt, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return 0, err
	}
	if len(result.Choices) == 0 {
		return 0, nil
	}

	raw := strings.TrimSpace(result.Choices[0].Message.Content)
	s.writeLog(requestID, nil, prompt, raw, time.Since(start).Milliseconds(), "success", "")

	var price int
	if _, err := fmt.Sscanf(raw, "%d", &price); err != nil {
		return 0, fmt.Errorf("parse price %q: %w", raw, err)
	}
	return price, nil
}

func (s *fireworksAIService) ParseOrders(requestID string, orders []ParseOrderInput, location string) ([]ParsedOrderItem, error) {
	if len(orders) == 0 {
		return []ParsedOrderItem{}, nil
	}
	start := time.Now()

	inputJSON, _ := json.Marshal(orders)
	input := string(inputJSON)

	locationCtx := ""
	if location != "" {
		locationCtx = fmt.Sprintf("Konteks lokasi: %s — sesuaikan estimasi harga dengan standar harga warung/kaki lima di sekitar lokasi tersebut.\n\n", location)
	}

	prompt := fmt.Sprintf(
		"Kamu adalah asisten parsing pesanan makan kolektif.\n\n"+
			"%s"+
			"Tugasmu:\n"+
			"1. Pisahkan pesanan majemuk menjadi item individual (misal 'Bakso + Nasi' → dua item terpisah; 'tempe tahu' → dua item: Tempe dan Tahu)\n"+
			"2. Keterangan khusus (misal 'ga pake sayur', 'sambal pedes', 'pake bihun', '(paha)', '(dada)') → sertakan dalam order_detail dalam tanda kurung, contoh: 'Ayam Cabe Ijo (Paha)'\n"+
			"3. Identifikasi vendor/lokasi beli dari konteks: nama orang+makanan (misal 'Bakso Pak Donan'), nama daerah (misal 'Ayam Madura'), atau lokasi spesifik (misal 'Rujak Tebet Depan BCA')\n"+
			"4. POLA WARTEG — 'Nasi [nama orang/warung]' (misal 'Nasi Bu Mul', 'Nasi Pak Slamet') → ini warung nasi/warteg:\n"+
			"   - vendor_name = nama orang/warungnya saja (misal 'Bu Mul', 'Pak Slamet')\n"+
			"   - item pertama = 'Nasi' (nasi putih dasar)\n"+
			"   - 'pake X + Y + Z' → setiap X, Y, Z adalah lauk terpisah\n"+
			"   - Harga lauk (jika beli DENGAN nasi): estimasi harga per-lauk warteg sesuai lokasi\n"+
			"   - Harga lauk (jika beli TANPA nasi, misal 'Bu Mul ayam goreng' saja): estimasi harga per-porsi warteg normal sesuai lokasi\n"+
			"5. POLA LAUK PENDAMPING — 'Nasi [jenis masakan]' (misal 'Nasi Ayam Madura', 'Nasi Cabe Ijo') → Nasi item terpisah di vendor yang sama\n"+
			"6. Pesanan dari vendor yang sama → vendor_name persis sama\n"+
			"7. Jika order_detail mengandung petunjuk harga eksplisit (misal '10rb', '10k', '15ribu', 'Rp 10.000', '10000') → gunakan angka tersebut sebagai estimated_price (rb/k = ×1000), hapus dari order_detail, JANGAN estimasi ulang\n"+
			"8. Jika tidak ada petunjuk harga → estimasi harga per item dalam Rupiah, wajar untuk warung/kaki lima Indonesia\n"+
			"9. Angka jumlah di akhir nama item (misal 'tempe goreng tepung 3', 'cimol 2') → ekstrak sebagai field qty, hapus angka dari order_detail. Angka qty BUKAN harga. Default qty = 1\n"+
			"10. Nama item: Title Case, singkat dan deskriptif, tanpa angka qty maupun petunjuk harga\n\n"+
			"Input JSON: [{\"name\":\"pemesan\",\"order_detail\":\"detail pesanan mentah\"}]\n"+
			"Output JSON: [{\"name\":\"pemesan\",\"order_detail\":\"nama item bersih\",\"vendor_name\":\"nama vendor\",\"estimated_price\":15000,\"qty\":1}]\n\n"+
			"Contoh:\n"+
			"Input: [{\"name\":\"Ali\",\"order_detail\":\"Nasi Bu Mul pake orek tempe + urap + tempe goreng tepung 3 + Tahu goreng\"},{\"name\":\"Silma\",\"order_detail\":\"cimol biasa bumbu campur pedes 10rb\"}]\n"+
			"Output: [{\"name\":\"Ali\",\"order_detail\":\"Nasi\",\"vendor_name\":\"Bu Mul\",\"estimated_price\":4000,\"qty\":1},{\"name\":\"Ali\",\"order_detail\":\"Orek Tempe\",\"vendor_name\":\"Bu Mul\",\"estimated_price\":4000,\"qty\":1},{\"name\":\"Ali\",\"order_detail\":\"Urap\",\"vendor_name\":\"Bu Mul\",\"estimated_price\":3000,\"qty\":1},{\"name\":\"Ali\",\"order_detail\":\"Tempe Goreng Tepung\",\"vendor_name\":\"Bu Mul\",\"estimated_price\":4000,\"qty\":3},{\"name\":\"Ali\",\"order_detail\":\"Tahu Goreng\",\"vendor_name\":\"Bu Mul\",\"estimated_price\":3000,\"qty\":1},{\"name\":\"Silma\",\"order_detail\":\"Cimol (Bumbu Campur Pedas)\",\"vendor_name\":\"Cimol\",\"estimated_price\":10000,\"qty\":1}]\n\n"+
			"Jawab JSON array saja, tanpa penjelasan, tanpa markdown.\n\n"+
			"Input: %s",
		locationCtx, input,
	)

	reqBody, _ := json.Marshal(map[string]any{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":       2048,
		"temperature":      0,
		"reasoning_effort": "none",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.fireworks.ai/inference/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		s.writeLog(requestID, nil, prompt, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.writeLog(requestID, nil, prompt, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("fireworks API %d: %s", resp.StatusCode, string(b))
		s.writeLog(requestID, nil, prompt, "", time.Since(start).Milliseconds(), "failed", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.writeLog(requestID, nil, prompt, "", time.Since(start).Milliseconds(), "failed", err.Error())
		return nil, err
	}
	if len(result.Choices) == 0 {
		s.writeLog(requestID, nil, prompt, "", time.Since(start).Milliseconds(), "success", "")
		return []ParsedOrderItem{}, nil
	}

	raw := strings.TrimSpace(result.Choices[0].Message.Content)
	if idx := strings.Index(raw, "["); idx > 0 {
		raw = raw[idx:]
	}
	if idx := strings.LastIndex(raw, "]"); idx >= 0 && idx < len(raw)-1 {
		raw = raw[:idx+1]
	}

	var items []ParsedOrderItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		s.writeLog(requestID, nil, prompt, raw, time.Since(start).Milliseconds(), "failed", err.Error())
		return nil, fmt.Errorf("parse response: %w", err)
	}

	s.writeLog(requestID, nil, prompt, raw, time.Since(start).Milliseconds(), "success", "")
	return items, nil
}

func (s *fireworksAIService) writeLog(requestID string, orderID *uint, input, output string, latencyMs int64, status, errMsg string) {
	if s.logRepo == nil {
		return
	}
	s.logRepo.Create(&models.AILog{
		RequestID: requestID,
		OrderID:   orderID,
		Input:     input,
		Output:    output,
		Model:     s.model,
		Provider:  "fireworks",
		LatencyMs: latencyMs,
		Status:    status,
		ErrorMsg:  errMsg,
	})
}
