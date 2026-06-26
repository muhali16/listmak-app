<template>
  <LegalLayout title="Syarat &amp; Ketentuan">
    <h2>Penerimaan</h2>
    <p>
      Dengan masuk dan menggunakan Listmak, Anda menyetujui Syarat &amp; Ketentuan
      ini beserta <router-link to="/privacy">Kebijakan Privasi</router-link> kami.
      Jika tidak setuju, mohon untuk tidak menggunakan aplikasi.
    </p>

    <h2>Tentang Layanan</h2>
    <p>
      Listmak adalah aplikasi untuk mencatat pesanan makan bareng, menghitung total
      secara otomatis, dan membagikan rekap melalui tautan. Layanan disediakan
      "sebagaimana adanya" dan dapat berubah seiring pengembangan.
    </p>

    <h2>Akun</h2>
    <ul>
      <li>Anda masuk menggunakan akun Google dan bertanggung jawab atas aktivitas pada akun tersebut.</li>
      <li>Anda wajib memberikan informasi yang benar dan menjaga keamanan akses Anda.</li>
    </ul>

    <h2>Penggunaan yang Wajar</h2>
    <p>Anda setuju untuk tidak:</p>
    <ul>
      <li>Menyalahgunakan layanan untuk aktivitas melanggar hukum atau merugikan orang lain.</li>
      <li>Mengganggu, membebani berlebih, atau mencoba membobol sistem.</li>
      <li>Menggunakan tautan berbagi untuk menyebarkan konten yang tidak pantas.</li>
    </ul>

    <h2>Konten Anda</h2>
    <p>
      Data pesanan, daftar, dan kontak yang Anda buat tetap menjadi milik Anda.
      Anda bertanggung jawab atas keakuratan data tersebut, termasuk perhitungan
      dan penyelesaian pembayaran antar anggota tim — Listmak hanya alat bantu
      pencatatan, bukan penyedia layanan pembayaran.
    </p>

    <h2>Ketersediaan</h2>
    <p>
      Kami berupaya menjaga layanan tetap tersedia, namun tidak menjamin bebas
      gangguan. Layanan dapat dihentikan sementara untuk pemeliharaan atau alasan
      teknis lain.
    </p>

    <h2>Batasan Tanggung Jawab</h2>
    <p>
      Sepanjang diizinkan hukum, enambelas.dev tidak bertanggung jawab atas
      kerugian tidak langsung yang timbul dari penggunaan layanan, termasuk
      kesalahan perhitungan atau kehilangan data.
    </p>

    <h2>Perubahan Ketentuan</h2>
    <p>
      Kami dapat memperbarui Syarat &amp; Ketentuan ini. Dengan terus menggunakan
      layanan setelah perubahan, Anda dianggap menyetujui versi terbaru.
    </p>

    <h2>Kontak</h2>
    <p>
      Listmak dibuat oleh <strong>enambelas.dev</strong>. Pertanyaan terkait
      ketentuan ini: <a href="mailto:hello@enambelas.dev">hello@enambelas.dev</a>.
    </p>
  </LegalLayout>
</template>

<script>
import LegalLayout from '../components/LegalLayout.vue'

export default {
  name: 'TermsView',
  components: { LegalLayout }
}
</script>
