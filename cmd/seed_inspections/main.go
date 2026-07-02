package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/ipincamp/radar-jentik-api/internal/adapters/driven/postgres"
	"github.com/ipincamp/radar-jentik-api/pkg/config"
	"golang.org/x/crypto/bcrypt"
)

type Point struct {
	HeadOfFamily string
	RT           int
	RW           int
	Village      string
	Lat          float64
	Lon          float64
	StatusJentik int
}

func main() {
	// 1. Load config and connect using shared Postgres helper (GORM)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Gagal load config: %v", err)
	}

	gormDB, err := postgres.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Gagal koneksi ke database: %v", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("Gagal mengambil sql.DB dari GORM: %v", err)
	}
	defer sqlDB.Close()

	rand.Seed(time.Now().UnixNano())

	// 2. Ambil Data Desa dari Database
	villageMap := make(map[string]string)
	rows, err := sqlDB.Query("SELECT id, name FROM villages")
	if err != nil {
		log.Fatalf("Gagal mengambil data villages: %v", err)
	}
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)
		villageMap[name] = id
	}
	rows.Close()

	// 3. Buat 5 Akun Kader (Satu per Desa)
	userMap := make(map[string]string)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	desaList := []string{"Cipete", "Kasegeran", "Panusupan", "Pejogol", "Langgongsari"}
	for _, desa := range desaList {
		villageID, ok := villageMap[desa]
		if !ok {
			log.Printf("Peringatan: Desa %s tidak ditemukan di tabel villages", desa)
			continue
		}

		username := fmt.Sprintf("kader_%s", time.Now().Format("150405")) // agar unique jika dijalankan ulang
		username = "kader_" + desa                                       // atau gunakan nama spesifik

		var userID string
		err := sqlDB.QueryRow(`
			INSERT INTO users (username, name, password, role, village_id) 
			VALUES ($1, $2, $3, 'cadre', $4) 
			ON CONFLICT (username) DO UPDATE SET updated_at = NOW()
			RETURNING id`,
			username, fmt.Sprintf("Kader %s", desa), string(hashedPassword), villageID).Scan(&userID)

		if err != nil {
			log.Fatalf("Gagal insert user kader %s: %v", desa, err)
		}
		userMap[desa] = userID
	}
	fmt.Println("✅ 5 Akun Kader berhasil dibuat (Password default: password123)")

	// 4. Ambil Master Tipe Kontainer
	var containerTypes []string
	cRows, _ := sqlDB.Query("SELECT id FROM container_types")
	for cRows.Next() {
		var id string
		cRows.Scan(&id)
		containerTypes = append(containerTypes, id)
	}
	cRows.Close()

	if len(containerTypes) == 0 {
		log.Fatal("Tabel container_types kosong! Jalankan seeder kontainer terlebih dahulu.")
	}

	// 5. Data 100 Titik Inspeksi
	points := get100Points()

	// 6. Mapping Bulan untuk masing-masing desa (Jan - Mei 2026)
	monthMap := map[string]time.Month{
		"Cipete":       time.January,
		"Kasegeran":    time.February,
		"Panusupan":    time.March,
		"Pejogol":      time.April,
		"Langgongsari": time.May,
	}

	// 7. Insert Inspection Reports dan Container Details
	count := 0
	for _, p := range points {
		villageID := villageMap[p.Village]
		userID := userMap[p.Village]
		if villageID == "" || userID == "" {
			continue
		}

		// Generate Tanggal Random berdasarkan bulan desa (Tahun 2026)
		month := monthMap[p.Village]
		day := rand.Intn(28) + 1 // 1-28 agar aman untuk Februari
		hour := rand.Intn(8) + 8 // Jam 08:00 - 16:00
		minute := rand.Intn(60)
		inspectionDate := time.Date(2026, month, day, hour, minute, 0, 0, time.Local)

		isPositive := p.StatusJentik == 1

		// Insert Report
		var reportID string
		err := sqlDB.QueryRow(`
			INSERT INTO inspection_reports 
			(user_id, village_id, family_head_name, rt, rw, latitude, longitude, larvae_status, photo_url, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
			userID, villageID, p.HeadOfFamily, fmt.Sprintf("%d", p.RT), fmt.Sprintf("%d", p.RW), p.Lat, p.Lon, p.StatusJentik, "", inspectionDate).Scan(&reportID)

		if err != nil {
			log.Printf("Gagal insert report untuk KK %s: %v", p.HeadOfFamily, err)
			continue
		}

		// Generate Container Details Realistis
		numContainers := rand.Intn(3) + 2 // Tiap rumah dicek 2-4 tipe wadah
		rand.Shuffle(len(containerTypes), func(i, j int) {
			containerTypes[i], containerTypes[j] = containerTypes[j], containerTypes[i]
		})
		selectedContainers := containerTypes[:numContainers]

		positiveDistributed := false
		for _, cTypeID := range selectedContainers {
			inspectedCount := rand.Intn(3) + 1 // Ditemukan 1-3 wadah per tipe
			positiveCount := 0

			if isPositive {
				if !positiveDistributed {
					positiveCount = rand.Intn(inspectedCount) + 1
					positiveDistributed = true
				} else {
					// Kadang-kadang ada kontainer lain yang juga positif
					if rand.Float32() > 0.7 {
						positiveCount = rand.Intn(inspectedCount) + 1
					}
				}
			}

			_, err = sqlDB.Exec(`
				INSERT INTO container_details (inspection_report_id, container_type_id, inspected_count, positive_count)
				VALUES ($1, $2, $3, $4)`,
				reportID, cTypeID, inspectedCount, positiveCount)

			if err != nil {
				log.Printf("Gagal insert container detail: %v", err)
			}
		}
		count++
	}

	fmt.Printf("✅ %d Data Inspeksi beserta Detail Kontainer berhasil di-seed!\n", count)
}

func get100Points() []Point {
	return []Point{
		// Cipete (20 Titik)
		{"Bpk. Supriyono", 1, 1, "Cipete", -7.421510, 109.132140, 0},
		{"Bpk. Mulyadi", 2, 1, "Cipete", -7.422105, 109.133250, 1},
		{"Ibu Kartini", 3, 1, "Cipete", -7.423402, 109.131550, 0},
		{"Bpk. Wahyudi", 1, 2, "Cipete", -7.424100, 109.134100, 1},
		{"Bpk. Agus S.", 2, 2, "Cipete", -7.420850, 109.135200, 0},
		{"Ibu Siti Aminah", 3, 2, "Cipete", -7.422750, 109.136050, 0},
		{"Bpk. Riyanto", 1, 3, "Cipete", -7.425120, 109.131100, 1},
		{"Bpk. Haryono", 2, 3, "Cipete", -7.423880, 109.130200, 0},
		{"Ibu Sumarni", 3, 3, "Cipete", -7.421990, 109.137300, 1},
		{"Bpk. Widodo", 4, 3, "Cipete", -7.424550, 109.135800, 0},
		{"Bpk. Herman", 2, 1, "Cipete", -7.421810, 109.132540, 0},
		{"Ibu Narti", 3, 1, "Cipete", -7.422505, 109.133650, 0},
		{"Bpk. Yudi", 4, 1, "Cipete", -7.423102, 109.131950, 1},
		{"Ibu Siska", 1, 2, "Cipete", -7.424500, 109.134500, 0},
		{"Bpk. Rizal", 2, 2, "Cipete", -7.420450, 109.135600, 0},
		{"Bpk. Tono", 3, 2, "Cipete", -7.422950, 109.136450, 1},
		{"Ibu Mimin", 4, 2, "Cipete", -7.425520, 109.131500, 0},
		{"Bpk. Surya", 1, 3, "Cipete", -7.423480, 109.130600, 0},
		{"Bpk. Arif", 2, 3, "Cipete", -7.421590, 109.137700, 1},
		{"Ibu Dewi", 3, 3, "Cipete", -7.424950, 109.135300, 0},

		// Kasegeran (20 Titik)
		{"Bpk. Sutrisno", 1, 1, "Kasegeran", -7.434100, 109.151200, 0},
		{"Ibu Rina W.", 2, 1, "Kasegeran", -7.435200, 109.152500, 1},
		{"Bpk. Joko P.", 3, 1, "Kasegeran", -7.433850, 109.153400, 0},
		{"Bpk. Ahmad M.", 1, 2, "Kasegeran", -7.436100, 109.150100, 0},
		{"Ibu Ningsih", 2, 2, "Kasegeran", -7.437250, 109.151800, 1},
		{"Bpk. Suryo", 3, 2, "Kasegeran", -7.432500, 109.154200, 0},
		{"Bpk. Budi H.", 1, 3, "Kasegeran", -7.438050, 109.153100, 1},
		{"Ibu Wulan", 2, 3, "Kasegeran", -7.435990, 109.155500, 0},
		{"Bpk. Tarno", 3, 3, "Kasegeran", -7.433120, 109.156100, 0},
		{"Bpk. Hendra", 4, 3, "Kasegeran", -7.437500, 109.154800, 1},
		{"Bpk. Taufik", 1, 1, "Kasegeran", -7.434500, 109.151600, 0},
		{"Ibu Sari", 2, 1, "Kasegeran", -7.435600, 109.152900, 0},
		{"Bpk. Lukman", 3, 1, "Kasegeran", -7.433450, 109.153800, 1},
		{"Bpk. Hasan", 4, 1, "Kasegeran", -7.436500, 109.150500, 0},
		{"Ibu Yulia", 1, 2, "Kasegeran", -7.437650, 109.152200, 0},
		{"Bpk. Dedi", 2, 2, "Kasegeran", -7.432900, 109.154600, 1},
		{"Bpk. Rahmat", 3, 2, "Kasegeran", -7.438450, 109.153500, 0},
		{"Ibu Lina", 1, 3, "Kasegeran", -7.435590, 109.155900, 1},
		{"Bpk. Wahyu", 2, 3, "Kasegeran", -7.433520, 109.156500, 0},
		{"Ibu Tuti", 3, 3, "Kasegeran", -7.437100, 109.154400, 0},

		// Panusupan (20 Titik)
		{"Bpk. Yanto", 1, 1, "Panusupan", -7.442100, 109.162300, 0},
		{"Ibu Endang", 2, 1, "Panusupan", -7.443500, 109.163100, 1},
		{"Bpk. Cipto", 3, 1, "Panusupan", -7.441200, 109.161500, 0},
		{"Bpk. Rudi H.", 1, 2, "Panusupan", -7.444800, 109.160900, 0},
		{"Ibu Sri Rahayu", 2, 2, "Panusupan", -7.445200, 109.162800, 1},
		{"Bpk. Darmo", 3, 2, "Panusupan", -7.442900, 109.164500, 1},
		{"Bpk. Firman", 1, 3, "Panusupan", -7.446100, 109.161200, 0},
		{"Ibu Sulastri", 2, 3, "Panusupan", -7.443800, 109.165100, 0},
		{"Bpk. Giyanto", 3, 3, "Panusupan", -7.441750, 109.166200, 1},
		{"Bpk. Heri", 4, 3, "Panusupan", -7.447050, 109.163900, 0},
		{"Bpk. Jaelani", 1, 1, "Panusupan", -7.442500, 109.162700, 0},
		{"Ibu Ratna", 2, 1, "Panusupan", -7.443900, 109.163500, 1},
		{"Bpk. Fajar", 3, 1, "Panusupan", -7.441600, 109.161900, 0},
		{"Bpk. Iwan", 4, 1, "Panusupan", -7.444400, 109.160500, 0},
		{"Ibu Maya", 1, 2, "Panusupan", -7.445600, 109.162400, 1},
		{"Bpk. Gunawan", 2, 2, "Panusupan", -7.442500, 109.164900, 0},
		{"Ibu Tari", 3, 2, "Panusupan", -7.446500, 109.161600, 0},
		{"Bpk. Pras", 1, 3, "Panusupan", -7.443400, 109.165500, 0},
		{"Bpk. Dwi", 2, 3, "Panusupan", -7.441350, 109.166600, 1},
		{"Ibu Rini", 3, 3, "Panusupan", -7.447450, 109.163500, 0},

		// Pejogol (20 Titik)
		{"Bpk. Karyo", 1, 1, "Pejogol", -7.428100, 109.165200, 0},
		{"Ibu Darti", 2, 1, "Pejogol", -7.429500, 109.166100, 1},
		{"Bpk. Lilik", 3, 1, "Pejogol", -7.427200, 109.164500, 0},
		{"Bpk. Sugiarto", 1, 2, "Pejogol", -7.430800, 109.167900, 0},
		{"Ibu Yanti", 2, 2, "Pejogol", -7.426500, 109.168800, 1},
		{"Bpk. Purnomo", 3, 2, "Pejogol", -7.428900, 109.169500, 0},
		{"Bpk. Slamet", 1, 3, "Pejogol", -7.431100, 109.165500, 1},
		{"Ibu Tini", 2, 3, "Pejogol", -7.427800, 109.170100, 0},
		{"Bpk. Bowo", 3, 3, "Pejogol", -7.425750, 109.167200, 1},
		{"Bpk. Andi", 4, 3, "Pejogol", -7.432050, 109.168200, 0},
		{"Bpk. Santo", 1, 1, "Pejogol", -7.428500, 109.165600, 0},
		{"Ibu Titin", 2, 1, "Pejogol", -7.429900, 109.166500, 0},
		{"Bpk. Budi", 3, 1, "Pejogol", -7.427600, 109.164900, 1},
		{"Bpk. Yos", 4, 1, "Pejogol", -7.430400, 109.167500, 0},
		{"Ibu Mega", 1, 2, "Pejogol", -7.426100, 109.168400, 0},
		{"Bpk. Narto", 2, 2, "Pejogol", -7.428500, 109.169900, 1},
		{"Ibu Lusi", 3, 2, "Pejogol", -7.431500, 109.165100, 0},
		{"Bpk. Karno", 1, 3, "Pejogol", -7.427400, 109.170500, 1},
		{"Bpk. Hari", 2, 3, "Pejogol", -7.425350, 109.167600, 0},
		{"Ibu Novi", 3, 3, "Pejogol", -7.432450, 109.168600, 0},

		// Langgongsari (20 Titik)
		{"Bpk. Maryanto", 1, 1, "Langgongsari", -7.412100, 109.148200, 0},
		{"Ibu Ratih", 2, 1, "Langgongsari", -7.413500, 109.149100, 1},
		{"Bpk. Kusno", 3, 1, "Langgongsari", -7.411200, 109.147500, 0},
		{"Bpk. Sigit", 1, 2, "Langgongsari", -7.414800, 109.150900, 1},
		{"Ibu Marni", 2, 2, "Langgongsari", -7.410500, 109.151800, 0},
		{"Bpk. Anton", 3, 2, "Langgongsari", -7.412900, 109.152500, 0},
		{"Bpk. Wahyo", 1, 3, "Langgongsari", -7.415100, 109.148500, 1},
		{"Ibu Retno", 2, 3, "Langgongsari", -7.411800, 109.153100, 0},
		{"Bpk. Dwi", 3, 3, "Langgongsari", -7.409750, 109.150200, 1},
		{"Bpk. Eko", 4, 3, "Langgongsari", -7.416050, 109.151200, 0},
		{"Bpk. Joko", 1, 1, "Langgongsari", -7.412500, 109.148600, 0},
		{"Ibu Rika", 2, 1, "Langgongsari", -7.413900, 109.149500, 1},
		{"Bpk. Harno", 3, 1, "Langgongsari", -7.411600, 109.147100, 0},
		{"Bpk. Deni", 4, 1, "Langgongsari", -7.414400, 109.150500, 0},
		{"Ibu Sinta", 1, 2, "Langgongsari", -7.410100, 109.151400, 1},
		{"Bpk. Bowo", 2, 2, "Langgongsari", -7.412500, 109.152900, 0},
		{"Ibu Nita", 3, 2, "Langgongsari", -7.415500, 109.148100, 0},
		{"Bpk. Tejo", 1, 3, "Langgongsari", -7.411400, 109.153500, 0},
		{"Bpk. Dimas", 2, 3, "Langgongsari", -7.409350, 109.150600, 1},
		{"Ibu Ayu", 3, 3, "Langgongsari", -7.416450, 109.151600, 0},
	}
}
