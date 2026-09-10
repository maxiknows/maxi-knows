# Project structure

```text
maxi-knows/
│
├── .github/
│
├── .idea/                          # lokal IDE-konfiguration - gitignored
│
├── .venv/                          # lokalt Python-miljø - gitignored
│
├── data/                           # database
│   ├── README.md
│   └── whoknows.db                 # gitignored
│
├── docs/                           # projektdokumentation
│   └── project-structure.md
│
├── go/                             # ny implementation i Go
│   ├── cmd/
│   │   └── maxi-knows/
│   │       └── main.go
│   │
│   ├── internal/
│   │   └── storage/
│   │       ├── init_db.go
│   │       └── queries.go
│   │
│   ├── go.mod
│   └── go.sum
│
├── legacy-python/                  # gammel legacy implementation
│   ├── backend/
│   │   ├── app.py
│   │   ├── app_tests.py
│   │   ├── requirements.txt
│   │   ├── static/
│   │   └── templates/
│   │
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
```
