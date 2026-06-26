<template>
  <LegalLayout title="Kebijakan Privasi">
    <h2>Ringkasan</h2>
    <p>
      Listmak adalah aplikasi untuk mengelola pesanan makan bareng tim. Kami hanya
      mengumpulkan data yang benar-benar dibutuhkan agar aplikasi berjalan. Kami
      <strong>tidak menjual</strong> data Anda ke pihak mana pun.
    </p>

    <h2>Data yang Kami Kumpulkan</h2>
    <ul>
      <li><strong>Identitas akun Google</strong> — nama, alamat email, dan foto profil, saat Anda masuk dengan Google.</li>
      <li><strong>Data pesanan</strong> — daftar (listmak), item pesanan, kontak tim, dan catatan yang Anda buat.</li>
      <li><strong>Data teknis</strong> — log permintaan dasar (waktu, jenis perangkat) untuk menjaga keandalan layanan.</li>
    </ul>

    <h2>Cara Kami Menggunakan Data</h2>
    <ul>
      <li>Mengautentikasi dan menjaga sesi Anda tetap masuk.</li>
      <li>Menyimpan dan menampilkan listmak serta pesanan yang Anda kelola.</li>
      <li>Membuat tautan berbagi agar tim dapat mengisi pesanan tanpa harus masuk.</li>
      <li>Menjaga keamanan, mencegah penyalahgunaan, dan memperbaiki layanan.</li>
    </ul>

    <h2>Berbagi Tautan (Share Link)</h2>
    <p>
      Saat Anda membagikan listmak, kami membuat tautan dengan token unik dan masa
      berlaku terbatas. Siapa pun yang memegang tautan tersebut dapat melihat dan
      mengisi pesanan pada listmak terkait. Jangan bagikan tautan ke pihak yang
      tidak Anda inginkan.
    </p>

    <h2>Penyimpanan & Keamanan</h2>
    <p>
      Sesi Anda disimpan melalui cookie aman (httpOnly). Sebagian data, seperti
      daftar pesanan harian, dapat disimpan secara lokal di peramban Anda agar
      aplikasi tetap cepat. Kami menerapkan langkah teknis yang wajar untuk
      melindungi data, namun tidak ada sistem yang sepenuhnya bebas risiko.
    </p>

    <h2>Layanan Pihak Ketiga</h2>
    <p>
      Kami menggunakan <strong>Google OAuth</strong> untuk proses masuk. Penggunaan
      Anda atas Google tunduk pada kebijakan privasi Google.
    </p>

    <h2>Hak Anda</h2>
    <p>
      Anda berhak mengakses, memperbaiki, atau meminta penghapusan data Anda.
      Untuk permintaan terkait data, hubungi kami di
      <a href="mailto:hello@enambelas.dev">hello@enambelas.dev</a>.
    </p>

    <h2>Perubahan</h2>
    <p>
      Kebijakan ini dapat diperbarui sewaktu-waktu. Tanggal "Terakhir diperbarui"
      di atas menunjukkan versi terbaru.
    </p>

    <h2>Kontak</h2>
    <p>
      Aplikasi ini dibuat dan dikelola oleh <strong>enambelas.dev</strong>.
      Pertanyaan seputar privasi: <a href="mailto:hello@enambelas.dev">hello@enambelas.dev</a>.
    </p>
  </LegalLayout>
</template>

<script>
import LegalLayout from '../components/LegalLayout.vue'

export default {
  name: 'PrivacyView',
  components: { LegalLayout }
}
</script>
