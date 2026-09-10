maxi-knows/
│
├── .github/
├── .idea/                          # lokal IDE-konfiguration - gitignored
├── .venv/                          # lokalt Python-miljø - gitignored
│
├── data/                           # database
│   └── whoknows.db                 # gitignored
│
├── docs/                           # projektdokumentation
│   └── project-structure.md
│
├── go/                             # ny implementation i Go
│   ├── cmd/
│   ├── internal/
│   ├── go.mod
│   └── go.sum
│
├── legacy-python/                  # gammel legacy implementation
│   ├── backend/
│   ├── backup_files/
│   ├── backend.py
│   ├── Makefile
│   ├── README.md
│   ├── run_forever.sh
│   └── schema.sql
│
├── .gitattributes
├── .gitignore
├── DEVELOPMENT_GUIDELINES.md
├── LICENSE
└── README.md