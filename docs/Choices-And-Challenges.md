# **Choices and Challenges**

Decisions, problems and thought processes from the Maxi Knows (WhoKnows) rewrite: from a Python 2 / Flask legacy app to a Go application.

**How to read this file**
- Entries are grouped by area. Each has a type (Decision / Challenge), who was involved, and the reasoning.
- ⚠️ **Missing** marks reasoning nobody wrote down. These are gaps to fill in as a group before the exam, not things to guess at.
- ⚠️ **Confirm** marks a statement that follows from the notes or commits but was not written down explicitly.
- 🔄 **Update** marks a change to an earlier entry. **Superseded** means the decision was replaced by a later one; the old entry stays because the reasoning still explains how we got here.
- Dates are 2026. People: Olga, Silke, Jihad, Tobias.

---

## **1. Architecture and repository**

### **Rewrite the legacy app in Go, standard library only**
- **Type:** Decision
- **Who:** Group (Olga, Silke, Jihad, Tobias)
- **When:** 4 Sep (framework decision), 4-11 Sep (prototype and rewrite start)
- **Context:** We inherited a Python 2 / Flask / Jinja app with SQLite. The rewrite language had to be chosen. Go and Ruby were the languages discussed in class.
- **Options considered:** Go with the standard library only; Go with a framework/router (Gorilla was raised); Ruby.
- **Outcome:** Go with the standard library. Pages use Go's `html/template`.
- **Why:** ⚠️ **Missing.** The 4 Sep notes say the group did not know how to choose between the standard library and Gorilla. What *is* documented is how we tried to learn enough to decide: before committing, we built a separate Go prototype on the branch `go-sqlite-setup` (Go module `whoknows`, a SQLite driver, `init_db.go` to create tables, `queries.go` for users and page search) to compare Go's syntax and usability with Python.
- **Trade-offs / lessons:** ⚠️ **Missing.** Also missing: why Go over Ruby, and why this SQLite driver.

### **One monorepo: `legacy-python/` and `go/` side by side**
- **Type:** Decision (later restructured)
- **Who:** Group
- **Context:** During the migration both the legacy and the new code must live somewhere, and the legacy app is our reference for behaviour.
- **Options considered:** Monorepo vs. multiple repos (documentation, legacy and Go in separate repos). Class notes list the monorepo arguments: transparency, and separate CI/CD pipelines per folder.
- **Outcome:** One repository.
- **Why:** The notes give the practical reason: both versions at the root, split by folder, so it is easy to see which code you are in.
- **Reversal:** The first layout kept the legacy app in `src/` next to `go/`. It was unclear which code was old and which was new, so the repo was restructured into `legacy-python/` (original app), `go/` (new app), `data/` (local database) and `docs/`.
- 🔄 **Update:** The Go code itself was later restructured into packages under `internal/` (see section 6, "Split `main.go` into packages").

### **Keep the SQLite database out of source code and out of Git**
- **Type:** Decision
- **Context:** Database files existed in more than one place in the repo, mixed with source code, so it was unclear which one the app used.
- **Outcome:** One `data/` directory for the local SQLite file, excluded via `.gitignore`.
- **Why:** Separates runtime data from code and removes the "which database is the real one" question.
- 🔄 **Update:** On the VM the database is now a mounted volume (`./data:/app/data`) used by the container, and the path can be set with the `DB_PATH` environment variable (default `../data/whoknows.db`). Tests use their own database built from `scripts/smoke-schema.sql`, so test data never touches the real one.
- **Open:** SQLite → Postgres was mentioned in class. ⚠️ **Missing:** was it considered for our project, and why was SQLite kept?

---

## **2. Workflow and collaboration**

### **GitHub Flow as branching strategy**
- **Type:** Decision
- **Who:** Group
- **Context:** The course required choosing a branching strategy (mandatory assignment, 18 Sep).
- **Options considered:** GitHub Flow; Git Flow (permanent `develop` and `release` branches).
- **Outcome:** GitHub Flow. Branch from `main` → name by purpose (`feat`, `fix`, `refactor`) → commit → push and open a PR using our PR template → review → merge → delete the branch.
- **Why:** It is simple, it gives one always-deployable `main`, and it fits a project our size. Git Flow added complexity we did not see a need for. (Jihad's write-up adds the same point from the safety side: separate branches keep damage, deletions and other risks away from the main version.)
- **Trade-offs / lessons:** A bad change merged to `main` could hit the deployable version, so review and testing before merge matter. Branches that live too long fall behind `main`; this *has* happened to us, so branches must stay short-lived. It needs good communication when several people touch related parts (we split login, weather API and linting across people).
- 🔄 **Update:** The flow now has automation attached. A PR triggers CI (Go tests, build, lint, SonarCloud analysis, Docker image build and smoke test). A merge to `main` publishes the image to GHCR and runs the deployment workflow (sections 3 and 4). That makes the "bad change reaches `main`" risk smaller, but see the next entry: nothing yet *forces* CI to pass before a merge.

### **Branch names, commit messages and PR/issue templates**
- **Type:** Decision
- **Who:** Group
- **Outcome:** Branch names are `<name>-<type>-<topic>` (e.g. `silke-feat-endpoints`). Commit types: `feat`, `fix`, `style`, `test`, `refactor`, `docs` (e.g. `feat: rewrite register page in Go`). A PR template (based on the course template) and an issue template are in the repo. Direct pushes to `main` are not allowed by agreement.
- **Why:** Course principle "work in small batches": smaller, more frequent commits, and a consistent PR process. ⚠️ **Missing:** why these specific types, and who proposed them.
- 🔄 **Update:** ⚠️ **Confirm:** the history uses `ci:` (e.g. `ci: add Hadolint Dockerfile linting`), which is not in the type list, and capitalisation varies (`Fix:` / `fix:`). Either add `ci` to `COMMIT_MESSAGE_RULES.md` or stop using it.

### **Branch protection rules: not enforced yet**
- **Type:** Open decision
- **Context:** The group asked: "can we stop people pushing to `main` without a PR?" Right now that depends on everyone remembering.
- **Status:** GitHub rulesets were identified as the tool. The notes say the group would ask AI what to define. ⚠️ **Missing:** whether the rules were created, and which.
- 🔄 **Update:** This matters more now that CI exists. Without a required-status-check rule, a PR with failing CI can still be merged. The only gate we know is real is inside the pipeline: `publish-image` will not run unless the smoke-test job is green. ⚠️ **Confirm:** a number of commits in the history have web-editor titles (`Update README.md`, `Create ...md`, `Rename ...`). Check whether these were made directly on `main`, since that would contradict the "no direct pushes" agreement.

### **GitHub organization, no license**
- **Type:** Decision
- **Outcome:** We created a GitHub organization, and chose not to add a license.
- **Why (organization):** Repositories can be shared as a group, instead of one person owning the repo and the others being added as collaborators.
- **Why (no license):** ⚠️ **Missing.**

### **SSH config alias for the VM (not a personal access token)**
- **Type:** Decision (convenience)
- **Who:** Olga (7 Oct)
- **Outcome:** Each developer can add a `Host` alias in their local SSH config so `ssh maxiknows` replaces the full command with IP, user and key.
- **Lessons:** On Windows, Notepad can save the file as `config.txt`, which SSH silently ignores ("Could not resolve hostname"). `ssh -G <alias>` shows whether the alias is being read. This is an SSH config, not a personal access token; PATs are for GitHub and the container registry.

---

## **3. Infrastructure and deployment**

### **Run the app on an Azure VM as a `systemd` service**
- **Type:** Decision, **superseded on 1 Nov** by the container setup below
- **Who:** Silke
- **When:** 24 Sep
- **Context:** The app was first started by hand on the VM (`go run`). It stopped whenever the SSH session ended.
- **Outcome:** A `systemd` unit that ran a built binary from `/usr/local/bin`, with `Restart=on-failure`.
- **Why:** The service starts automatically and restarts itself if the process fails.
- **Why Azure and this region:** ⚠️ **Missing.**

### **Challenge: `systemd` refused to start the binary (SELinux)**
- **Type:** Challenge
- **Who:** Silke
- **Symptom:** The service failed with `status=203/EXEC`, but the program ran fine when started manually, so the Go code was not at fault.
- **Diagnosis:** The binary was built inside the project folder and got the SELinux type `user_tmp_t`. SELinux does not let `systemd` execute files with that type.
- **Fix:** Copy the built binary to `/usr/local/bin/maxi-knows`, where it gets the type `bin_t`.
- **Lesson:** "Works manually, fails under systemd" is a hint to check SELinux and file context.
- 🔄 **Update:** SELinux came up again with HTTPS (below). This time nothing had to be changed: the certificates were readable by the Nginx container through a read-only volume mount.

### **Manual deployment process (historic)**
- **Type:** Decision, **superseded**
- **Who:** Silke
- **Process:** SSH to VM → `git switch main` → `git pull origin main` → `cd go` → `go build -o maxi-knows ./cmd/maxi-knows` → `sudo systemctl stop maxi-knows` → `sudo cp maxi-knows /usr/local/bin/maxi-knows` → `sudo systemctl restart maxi-knows` → `systemctl status maxi-knows`. The service must be stopped before copying, otherwise the binary is in use.
- 🔄 **Update:** Replaced by the container and the deployment workflow (next entry and section 4). The VM no longer builds anything; it pulls a finished image.

### **Move the VM from a `systemd` binary to a Docker container**
- **Type:** Decision with two challenges
- **Who:** Tobias (cutover, 1 Nov); Silke (Dockerfile and Compose, see section 4)
- **Context:** The app was already being built into a Docker image in CI. Running a different artifact on the VM (a binary built there) than the one we test in CI meant the tested thing and the deployed thing could differ.
- **Outcome:** The old `maxi-knows.service` was stopped and the app now runs from the GHCR image via `docker compose`.
- **Challenge 1: port clash.** The old service still held port 8080, so the container could not start. Fix: stop the service first, then `docker compose down` and `up -d`.
- **Challenge 2: "readonly database".** The container runs as a non-root user (uid 10001), but `data/` was owned by the VM user. Fix: `chown -R 10001:10001 data`. This is the same reason CI uses `chown 10001` instead of `chmod 777`.
- **Consequences:** `data/` is now owned by uid 10001, so `sqlite3` on the VM needs `sudo`. Anything that reads the database from outside the container (backups, see below) has to be checked against this.
- ⚠️ **Confirm (not yet checked in the notes):** the old service must be disabled so it does not grab port 8080 on reboot (`systemctl is-enabled maxi-knows.service`), Docker must start on boot (`systemctl is-enabled docker`), and a real VM reboot should be tested.
- ⚠️ **Missing:** why a container on the VM rather than keeping the binary. The reason above (tested artifact = deployed artifact) is the obvious one but is not written down by the group.

### **Domain, HTTPS and an Nginx reverse proxy**
- **Type:** Decision with several challenges
- **Who:** Olga
- **When:** 4 Nov
- **Context:** The app was reachable on the VM's IP and port 8080 over plain HTTP. Two existing features needed HTTPS: the weather widget uses the browser geolocation API, which requires a secure context, and the login cookie is set with `Secure: true`, which browsers only send over HTTPS.
- **Outcome:**
  - Domain `maxi-knows.dk` (registered at one.com, activation and ID verification through Punktum.dk) with A-records for the bare domain and `www`.
  - Ports 80 and 443 opened in both the Azure network security group and the AlmaLinux firewall; 22 stays open for SSH.
  - A Let's Encrypt certificate from Certbot, run in a temporary Docker container in standalone mode. Stored on the VM, not in Git. Expires 2 Feb 2027.
  - Nginx as a separate container: redirects HTTP to HTTPS and forwards to the Go app on 8080. The app's 8080 is now `expose`d inside the Compose network only, so Nginx is the only thing publishing ports on the VM. Config is in `nginx/nginx.conf`.
  - The deployment workflow now copies `compose.yaml` and `nginx/nginx.conf` to the VM, and the production smoke test uses `https://maxi-knows.dk`. CI still tests the app directly on 8080 in its own environment.
- **Why Nginx as a reverse proxy in front of the app:** It keeps TLS out of the Go code and keeps the app off the public network. ⚠️ **Confirm:** the notes explain why HTTPS was needed, not why Nginx over alternatives.
- **Challenges:**
  - *DNS:* NXDOMAIN and SERVFAIL while testing. Google's resolver returned SERVFAIL while one.com's returned the right IP; Cloudflare resolved it later. Checked the A-records, nameservers and domain activation, then waited for propagation.
  - *`nginx -t` failed locally* because the standalone test container was not on the Compose network and could not resolve the hostname `maxi-knows`. Rerun on the VM inside the `maxi-knows_default` network, where it passed.
  - *SELinux:* confirmed the certificates were readable through a read-only mount. No policy change needed.
- **Trade-off, no automatic certificate renewal:** decided deliberately, because the app is planned to be shut down after the exam, before 2 Feb 2027. Standalone-mode renewal needs port 80, which Nginx now holds, so renewing later would mean stopping Nginx first.
- **Verified:** site loads over HTTPS, weather widget gets location and weather, login and logout work, both containers run, only Nginx publishes HTTP/HTTPS.
- **To do after the exam:** remove the old Azure inbound rule for 8080 if it is still there, shut down the Azure resources, and switch off domain auto-renewal to avoid renewal costs.

### **Fail2Ban**
- **Type:** Decision
- **Who:** Olga (4 Nov)
- **Outcome:** Documented in `docs/fail2ban.md`. Read that file for what is configured.
- ⚠️ **Missing:** why we added it and which threat the group had in mind. The commit history and logs only record that it was done.

### **The VM was replaced (Spain Central → France Central)**
- **Type:** Decision (infrastructure change)
- **When:** around 21 Oct (the logs mention "the new Azure VM")
- **What changed:** New VM `maxi-knows-vm` in France Central with a new public IP and login user, replacing `vm1-alma` in Spain Central. Same OS (AlmaLinux with SELinux).
- **Consequence:** The database was restored on 21 Oct from the 7 Oct backup (Olga): stop service, copy file over `scp`, `PRAGMA integrity_check` returned `ok`, replace `whoknows.db`, restart, check the journal. Anything configured on the old VM (cron jobs, scripts, firewall rules) had to be redone, which is why the backup scripts below need checking.
- ⚠️ **Missing:** why the VM was replaced and why the region changed. Also still missing from before: why Azure at all.

### **Challenge: a feature branch ended up on the VM**
- **Type:** Challenge, **probably resolved**
- **When:** 30 Sep
- **Context:** After the weather API was implemented (Tobias), the deployment apparently pushed his feature branch straight to the VM. We only want `main` deployed.
- **Status:** The VM no longer runs code from a Git checkout; it runs the image published from CI. The deployment workflow also had a fix named "use trusted main branch for deployment checkout" (Jihad). ⚠️ **Confirm:** that this closes the original problem, and that image publishing only happens from `main`.
- **Follow-up:** The VM returned "weather unavailable" at the time, which looked like a permissions problem. It works now (verified on the HTTPS site on 4 Nov). ⚠️ **Missing:** the actual root cause. The commit history only says a Dockerfile fix was made "to make weather work. hopefully", and a HTTPS workaround for the weather service was committed on 30 Sep. Neither says what was wrong.

### **Challenge: the VM silently created an empty database**
- **Type:** Challenge, **resolved**
- **Who:** Silke (found, 20 Sep); Tobias (fixed, 1 Nov)
- **Context:** The VM needed access to the real data.
- **Diagnosis:** The Go program has a database path. When no file exists there, it creates an empty database instead of failing. The VM had therefore ended up with an empty `whoknows.db`.
- **Fix at the time:** Copied our real database to the VM with `scp`, and renamed the empty file to keep it.
- 🔄 **Update:** The program now stops at startup with a clear error if the database file is missing. The group question "should a missing file be an error?" is answered: yes. The same cleanup added a `CREATE TABLE IF NOT EXISTS sessions` at startup (see section 5) and an `http.Server` with timeouts.

### **Database backups**
- **Type:** Decision (several overlapping mechanisms), needs a status check
- **2-6 Oct, manual (Silke):** SSH into the VM, copy the database, `scp` to a local machine, upload to the team Google Drive.
- **7 Oct, automated on the VM (Olga):** a script runs daily at 02:00 UTC using SQLite's `.backup` (not a plain file copy, so it is consistent while the app is writing), runs an integrity check on the result, and deletes backups on the VM after 14 days. It can also be run by hand before a big deploy.
- **7 Oct, weekly local copy (Olga):** a PowerShell script on her Windows machine downloads the newest VM backup via `scp`, skips it if it already exists locally, runs weekly through Task Scheduler, and keeps 8 weeks.
- **26 Oct, daily local copy (Jihad):** a second Windows setup, run at login through Task Scheduler, at most one backup per calendar day. It reads the live database on the VM over SSH, makes an in-memory snapshot with Python's `sqlite3` backup API, rebuilds it locally, then checks `PRAGMA integrity_check` and that the `users` and `pages` tables exist. It never overwrites an existing file.
- **Why:** two copies on two machines, so losing the VM does not lose the data, and the integrity checks mean a corrupt backup is caught when it is made, not when it is needed.
- ⚠️ **Confirm (important):**
  - Olga's VM-side script and cron job were set up on the old VM. Is the 02:00 job also running on the new VM?
  - Since 1 Nov `data/` is owned by uid 10001. Check that Jihad's script, which reads `whoknows.db` as the normal VM user, still produces a valid backup (restore one into a scratch file and open it).
  - There are now three mechanisms. Decide which are the official ones and drop the rest.

---

## **4. CI/CD and testing**

### **Dockerize the app**
- **Type:** Decision
- **Who:** Silke (Dockerfile, 7 Oct), Tobias (`.dockerignore`)
- **Outcome:** Dockerfile and `.dockerignore`, with the container running as a non-root user (uid 10001). Environment files are excluded from the build context.
- **Why:** The non-root and no-env-files choices are security hygiene. ⚠️ **Missing:** why Docker at all, beyond that it was a course topic (the logs list "Docker" as a task on 30 Sep), and what the "Docker conflicting with the linter" fix on 8 Oct was.
- **Consequence:** The non-root user caused the "readonly database" problem on the VM (section 3).

### **Docker Compose and publishing the image to GHCR**
- **Type:** Decision
- **Who:** Silke (Compose), Olga (publish step)
- **When:** 26-28 Oct
- **Outcome:** `compose.yaml` describes the app (and later Nginx). CI builds the image and publishes it to GitHub Container Registry. Compose on the VM pulls the published image instead of building.
- **Why:** The VM should run exactly what CI built and tested. ⚠️ **Missing:** why GHCR rather than Docker Hub or another registry.

### **GitHub Actions CI**
- **Type:** Decision
- **Who:** Olga (workflow, 26 Oct, PR #100), Tobias (smoke-test job)
- **Outcome:** On PRs and pushes to `main`: run Go tests, build the app, build the Docker image. The Docker job only runs if the Go job passes. A `smoke-test` job (below) starts the built image and tests it. `publish-image` requires that job to be green.
- **Why test the image and not just the binary:** Mistakes in the Dockerfile and in file paths are caught before anything is published.
- **Hardening along the way:** `persist-credentials: false` on checkout, and third-party actions pinned to full commit SHAs (SonarCloud raised SHA warnings; pinning followed for the Docker actions on 28 Oct).
- **Trade-offs / lessons:** There are still no automated Go unit tests (`go test ./...` passes but covers little). The pipeline tests behaviour from the outside instead (smoke test, Playwright).

### **Go linting with golangci-lint**
- **Type:** Decision
- **Who:** Jihad (30 Sep)
- **Outcome:** A lint workflow in GitHub Actions; the action is pinned to a full commit SHA.
- **Why:** Part of the course's software-quality / linting / CI assignment. This closes the earlier "linting is still TODO" note.
- ⚠️ **Missing:** why golangci-lint and which linters are enabled. Silke asked earlier how this differs from SonarCloud (section 6); a one-line answer belongs here (linting = style and common bugs, SonarCloud = security and maintainability).

### **Lint Dockerfiles with Hadolint**
- **Type:** Decision
- **Who:** Silke (9 Nov)
- **Outcome:** Hadolint runs in CI on the Dockerfile; the action is pinned to a commit SHA, same as the others.
- ⚠️ **Missing:** what prompted it (a finding, or the course) and whether it fails the build or only warns.

### **Deployment workflow (CD)**
- **Type:** Decision
- **Who:** Jihad (workflow and SSH test), Tobias (checkout and smoke step)
- **When:** 28 Oct to 4 Nov
- **Outcome:** A GitHub Actions workflow connects to the VM over SSH using credentials stored in GitHub secrets (a dedicated deploy key). It checks out the trusted `main` branch, copies `compose.yaml` and `nginx/nginx.conf` to the VM, runs `docker compose pull` and `docker compose up -d`, then runs the smoke test against `https://maxi-knows.dk`.
- **Why the production smoke test is read-only:** it runs without `FULL=1`, so it does not register test users in the production database.
- **Challenge:** An SSH connection test came first (28 Oct) to prove the runner could reach the VM at all, before any deploy logic was added.
- ⚠️ **Confirm:** the trigger. Jihad's write-up says a merge to `main` triggers it. Tobias's 1 Nov entry still described the deploy step as manual (`docker compose pull && up -d` on the VM), and the later changes add the pull/up step. State which it is now, and whether `workflow_dispatch` is still available.
- ⚠️ **Missing:** why SSH-based deployment rather than another approach.

### **Smoke test**
- **Type:** Decision (optional course assignment)
- **Who:** Tobias
- **Outcome:** `scripts/smoke.sh` checks the pages, search, and weather/location, and with `FULL=1` also register, login and logout, including that an old session is invalid after logout. It exits 1 on any failure. `scripts/smoke-schema.sql` creates an empty database for CI. In CI it runs against the Docker image; in CD it runs against production, read-only.
- **Why:** One script that works as a CI gate and as a post-deploy check.
- **Sonar findings fixed on the way:** `[` → `[[` in the script, `chmod 777/666` → `chown 10001` in CI, and the MD5 verification marked `NOSONAR` with the reason (legacy hashes only, see section 5).

### **Playwright end-to-end tests**
- **Type:** Decision
- **Who:** Olga (9 Nov)
- **Outcome:** Playwright with Chromium, started from the teacher's Flask example and adapted to our Go app. Tests live in `go/tests/e2e/`: homepage and search elements, a search, and register/login/logout. A separate SQLite database is built from `scripts/smoke-schema.sql` and selected with `DB_PATH`. A `playwright.yml` workflow runs them on PRs to `main` and pushes to `main` and uploads the report. Details are in `docs/playwright-e2e-testing.md`.
- **Challenge:** SonarQube's security recommendations for the workflow were addressed (including a fix for how the Playwright dependencies are installed).
- ⚠️ **Missing:** why both a smoke test and Playwright. They overlap on pages, search and auth. A fair answer is that Playwright drives a real browser (forms, redirects, cookies) and the smoke test is a fast HTTP check that also runs in production, but the group should say so.

---

## **5. Security**

### **Replace MD5 password hashing with bcrypt, with a gradual migration**
- **Type:** Decision
- **Who:** Olga (branch `olga-fix-password-security`)
- **When:** 29 Sep
- **Context:** SonarCloud flagged MD5 for password storage in `main.go`. MD5 is old and weak. But the database already held users from the course simulation with MD5 hashes.
- **Options considered:** Switch to bcrypt immediately (existing users could no longer log in); or migrate gradually.
- **Outcome:** New registrations use bcrypt. On login, the app checks whether the stored hash is bcrypt. If it is an old MD5 hash and the password is correct, the password is re-hashed with bcrypt and the database is updated on the spot.
- **Why:** Keeps the existing simulation users working while moving everyone to bcrypt without a forced reset.
- **Trade-offs / lessons:** Users who never log in again keep their MD5 hash, and the MD5 verification path has to stay until then. If that matters, set a deadline after which MD5 login is removed.
- 🔄 **Update:** The MD5 check now lives in `internal/auth/password.go`, uses a constant-time comparison, and carries a `NOSONAR` comment explaining it is only used to verify old hashes. The old hash is upgraded at first login. The "never logs in" gap is still open.

### **Remove `DROP TABLE IF EXISTS users` and the hardcoded admin from `init_db.go`**
- **Type:** Decision, found while fixing the above
- **Who:** Olga
- **Context:** `init_db.go` dropped the `users` table on every run, so re-running initialization would delete all users. It also created an admin user with an MD5 hash.
- **Outcome:** Tables are created with `CREATE TABLE IF NOT EXISTS`. The hardcoded admin is gone; users are created through normal registration (bcrypt).

### **Session-based authentication** (was: "Known limitation: no sessions")
- **Type:** Challenge, **resolved 21 Oct**
- **Who:** Olga (branch `olga-feat-session-management`)
- **Finding (23 Sep and 21 Oct):** Login validated the password but issued no session or token, so the app could not show who was logged in and `/api/logout` had nothing to end. A Postman test confirmed it: correct status codes everywhere, but no `Set-Cookie` and no real session invalidation.
- **Decision:** Recreate what the Python app did with Flask sessions, in Go.
- **Outcome:**
  - A `sessions` table in SQLite, with functions to create, fetch and delete a session.
  - On login a random token from `crypto/rand`; only its SHA-256 hash is stored in the database, so a leaked database does not leak usable tokens.
  - The token is sent as a cookie with `HttpOnly`, `SameSite=Lax`, 24-hour expiry. On 26 Oct this was tightened to `Secure` (secure session cookies enforced), which is one reason HTTPS became necessary.
  - Page handlers read the cookie and show the username in the navigation.
  - Logout changed from GET to POST (a form and button instead of a link). It deletes the session in the database and clears the cookie.
  - Added `http.CrossOriginProtection` against unwanted cross-origin requests.
- **Why server-side sessions:** matches the legacy behaviour, and logout becomes real: a deleted session cannot be reused. ⚠️ **Confirm:** the notes do not compare this with signed or stateless tokens.
- **Tested:** Postman, DataGrip and browser. Active sessions in SQLite dropped from 3 to 2 after a logout. A request with a foreign origin was rejected with 403 (the first attempt returned 303 only because the server had not been restarted). `smoke.sh` with `FULL=1` also checks that the old session is invalid after logout, which covers the "deleted token cannot be reused" item from Olga's to-do list.
- **Challenges:**
  - *Compile errors from a function placed in the wrong `session.go` and a missing `}`.* Fixed.
  - *Login returned 500 "Could not save session" on a clean database.* The sessions code inserted into a table that nothing created, so it only worked on databases where someone had made the table by hand. Fix (Tobias, 1 Nov): `main.go` runs `CREATE TABLE IF NOT EXISTS sessions` on startup.
  - *Logout as a form broke the navigation layout* until the CSS was adjusted.
- **Still open from Olga's list:** automated Go tests for the session code, and updating the API documentation for the logout change (see the Swaggo entry).

---

## **6. Legacy code and software quality**

### **Modernize the legacy app to Python 3 first, then rewrite**
- **Type:** Decision and challenge
- **Who:** Group
- **When:** 28 Aug (upgrade assignment)
- **Context:** The legacy app needed to run before we could understand it well enough to rewrite it.
- **Challenges and fixes:**
  - *Pinned dependencies* (`Flask==0.5`, `Jinja2==2.4`, `Werkzeug==0.6.1`, `setuptools==44.1.1`) could not be installed on a modern Python. Fix: remove the version pins after guidance from the teacher.
  - *`2to3` did not work on Python 3.14*, because `lib2to3` was removed. Fix: install Python 3.12 with `uv` and recreate the virtual environment.
  - `2to3` converted `app.py` (print statements, a `__future__` import) and made a `.bak` backup, which is now git-ignored along with `backup_files/`.
- **Why an intermediate step:** The original could not run on a supported Python. The Go rewrite removes the Python 2 dependency permanently, while the Python 3 version is the working reference in the meantime.
- **Trade-offs / lessons:** Removing pins means unpinned, newer versions. This is acceptable for a throwaway reference app, but would not be for production.

### **Clean up the legacy scripts: Makefile, `run_forever.sh`, line endings**
- **Type:** Decision and challenge
- **Context:** The Makefile and shell script called `python2`. ShellCheck found an unquoted path, an unused `TMP` variable, and exit-code handling that lost the real exit code. The project is developed partly on Windows, so shell scripts had CRLF line endings, which break on Linux.
- **Outcome:** `python2` → `python3`; path quoted; exit code saved immediately (`exit_code=$?`); ShellCheck now reports nothing. `.gitattributes` forces LF: `*.sh text eol=lf` and `Makefile text eol=lf`.
- **Why:** Same behaviour on every developer's OS.

### **Legacy tests were not fixed**
- **Type:** Decision to defer
- **Context:** `app_tests.py` fails in `tearDown()` (`unlink()` on a temp file object does not exist in Python 3). The failure is in cleanup, so it does not necessarily mean login, register or search are broken.
- **Outcome:** Documented, not fixed. Fixing it was not part of the assignment. The tests are to be replaced as part of the rewrite.
- 🔄 **Update:** The replacement is the smoke test and the Playwright tests (section 4), not a port of `app_tests.py`.

### **Use SonarCloud and Qlty for software quality; ignore the legacy code**
- **Type:** Decision
- **Who:** Olga
- **When:** 24-30 Sep
- **Options considered:** SonarCloud (SonarQube), Code Climate / Qlty, CodeRabbit, DeepSource.
- **Outcome:** SonarCloud and Qlty are connected to the repo. Issues SonarCloud reports in `legacy-python/` are marked as ignored instead of fixed.
- **Why:** Olga's view: SonarCloud focuses on things like security, rather than the formatting a linter catches. Silke asked whether this overlaps with linting and whether we are allowed to change legacy code at all. ⚠️ **Missing:** why these two tools out of the four, and the explicit reason for ignoring legacy findings (most likely that it is being replaced, but not written down).
- 🔄 **Update:** Sonar findings have since driven real changes: bcrypt, form labels, the pinned action SHAs, `chown` instead of `chmod 777`, `[[` in shell, and the `NOSONAR` on the legacy MD5 check. That is the strongest evidence the tool earned its place.

### **Refactor the Go storage package (`main redeclared`)**
- **Type:** Challenge, with a reversal
- **Who:** Olga
- **When:** 29 Sep
- **Context:** `init_db.go` and `queries.go` both lived in `internal/storage` and each had its own `func main()`. Early on we noted this and moved on, since they were meant as two standalone programs and each ran fine alone. Qlty then failed the build with `main redeclared in this block`.
- **Diagnosis:** In Go, all files in one directory are one package, and a package can have only one `main()`.
- **Fix:** `internal/storage` is now `package storage` with no `main()`. Initialization became `InitDB(db *sql.DB)`. A new program in `go/cmd/init-db/main.go` opens the database and calls it. `queries.go` was kept for the planned consolidation of queries.
- **Result:** The Qlty workflow passes. Two typecheck warnings about imports of `internal/storage` and `modernc.org/sqlite` remain. They appear to be Qlty's Go configuration, not the code, since it builds locally. Still being investigated.
- **Lesson:** `cmd/` holds runnable programs; `internal/` holds reusable code without `main()`.

### **Split `main.go` into packages**
- **Type:** Decision with bug fixes found on the way
- **Who:** Tobias (branches `tobi-fix-main.go-cleanup` and `tobi-feat-smoketest`, PR #111, 1 Nov)
- **Context:** `main.go` had grown to about 500 lines with handlers, SQL, caching and external API calls mixed together.
- **Outcome:** `internal/geo` (weather and location lookups, cache, coordinate validation), `internal/auth` (password checks with bcrypt and legacy MD5, session helpers), `internal/storage/users.go` (all user SQL), `internal/handlers` (all handlers and routes). `main.go` only wires up the database, handlers and server.
- **Why:** One place per responsibility; the weather and location cache logic existed twice and now exists once.
- **Bugs found while moving code:**
  - *Data race:* handlers used `err =` instead of `err :=`, so one `err` variable was shared between all concurrent requests.
  - *Registration race:* two simultaneous registrations of the same name now return `ErrUserExists` instead of failing in an undefined way.
  - *Legacy MD5 comparison* is now constant-time.
- **Known tech debt left in (Tobias's list):** handlers return 500 without logging the real error (add `log.Printf` before `http.Error`), and templates are parsed on every request (parse once at startup or use `go:embed`).
- **Lesson:** Splitting a file is also a code review. Three real bugs surfaced because the code had to be read line by line.

### **Add form labels to the Go templates (accessibility)**
- **Type:** Decision
- **Who:** Olga
- **Context:** SonarCloud flagged inputs without `<label>` in `login.html`, `register.html` and `search.html`.
- **Outcome:** Labels linked via `for`/`id`. `name` attributes were untouched, so the forms send the same values to the backend.

---

## **7. Features and API**

### **Weather: Open-Meteo for forecasts, Nominatim for place names, with caching**
- **Type:** Decision
- **Who:** Tobias (implementation)
- **Context:** The course's "consumer report" assignment required a weather page backed by an external API.
- **Outcome:** `GET /api/weather` proxies Open-Meteo, with responses cached for 15 minutes. `GET /api/location` reverse-geocodes via Nominatim, cached for 24 hours. In both, coordinates are rounded to 2 decimals and an `X-Cache` header reports HIT or MISS.
- **Why:** ⚠️ **Missing:** why these two APIs, why proxy them through our backend, and why the cache durations and rounding. (Rounding to 2 decimals means nearby requests share a cache entry, which is probably the point, but confirm.)
- **Challenges:** Silke raised that the weather endpoint must work for the course session, and Olga (24 Sep) noted it was hard to know what "done" meant for weather. On the VM, the endpoint returned "weather unavailable" (section 3). Once the site moved to HTTPS the browser geolocation worked, which the widget needs.
- 🔄 **Update:** The code now lives in `internal/geo` as `cache.go`, `client.go`, `coords.go` and `service.go`, with one shared cache implementation.

### **Generate the API spec with Swaggo (Swagger 2.0)** — reversed
- **Type:** Decision, **later removed (29 Oct)**
- **Who:** Group; removal by Tobias
- **Context:** Mandatory assignment: produce an OpenAPI specification for our application.
- **Outcome at the time:** Handlers were annotated with Swaggo comments, and `swagger.json` and `swagger.yaml` were generated from them. The result covered `/api/search`, `/api/login`, `/api/register`, `/api/logout`, `/api/weather` and `/api/location`.
- **Why it was chosen:** The spec can be generated from code rather than written by hand, so it stays in sync, and it can be diffed against the spec handed out in class.
- **Trade-offs / lessons:** The generated file declares `"swagger": "2.0"`, not OpenAPI 3.x, which is what the course material discusses.
- 🔄 **Update:** The Swaggo docs were removed because they were "redundant". The old spec is kept on the branch `tobi-feat-swaggo` if a new one is needed. ⚠️ **Missing:** what made it redundant. A likely candidate is that the generated spec had drifted (logout moved from GET to POST and sessions were added on 21 Oct, and Olga's notes list "update the Swagger docs" as open), but that is a guess, not something the group wrote down. ⚠️ **Confirm:** where the spec that satisfies the mandatory assignment now lives.

### **Monitor the API with Postman**
- **Type:** Decision and challenge
- **Who:** Silke
- **When:** 6 Oct
- **Outcome:** Two tests on `GET /api/search?q=JavaScript`: status is 200 and the body is not empty. The monitor runs hourly.
- **Challenge 1:** Postman only accepted tests on *all* endpoints of a collection when creating a monitor, so a new collection with a single endpoint was made.
- **Challenge 2:** Sharing the monitor link requires a paid plan. The free plan allows about 1000 requests/month; hourly runs use about 720/month, which is close to the limit.
- **Open:** One shared monitor owned by Silke, or one each? Not decided (separate monitors would multiply request usage against the limits).
- 🔄 **Update:** ⚠️ **Confirm:** the monitor was set up before the VM was replaced and before port 8080 stopped being published. If it still targets the old IP and port, it is failing or monitoring a dead host. Point it at `https://maxi-knows.dk` and check that it is green.

---

## **8. Open items (need an answer from the group)**

| # | Item | Who could answer |
|---|------|------------------|
| 1 | Why Go over Ruby, and why the standard library over Gorilla | Everyone |
| 2 | Are branch protection rules / required status checks enabled? Were any commits made directly on `main`? | Whoever has repo admin |
| 3 | Was the "feature branch on the VM" problem closed by the trusted-main checkout and image-based deploy? | Jihad, Tobias |
| 4 | Root cause of the old "weather unavailable" error (it works now) | Tobias |
| 5 | Why the weather APIs and cache settings? | Tobias |
| 6 | Why SonarCloud and Qlty, and why ignore legacy findings? | Olga |
| 7 | Backups: which of the three mechanisms are live on the new VM, and do they still work with `data/` owned by uid 10001? Restore one to prove it | Olga, Jihad, Silke |
| 8 | Old `maxi-knows` systemd service disabled, Docker enabled on boot, reboot tested? | Tobias |
| 9 | Retire MD5 login after a deadline? | Olga |
| 10 | Why was Swaggo dropped, and where is the spec for the assignment? | Tobias |
| 11 | Postman monitor: ownership, and is it pointed at the new URL? | Silke |
| 12 | Why Azure, why the VM was replaced and moved region, why no license | Group |
| 13 | Why Docker, GHCR, Nginx, Fail2Ban and golangci-lint/Hadolint; why both a smoke test and Playwright | Olga, Silke, Jihad |
| 14 | CD trigger: automatic on merge to `main`, or manual? | Jihad |
| 15 | Tech debt: handlers log nothing on 500; templates parsed on every request; no Go unit tests | Tobias |
| 16 | After the exam: close the old 8080 inbound rule, shut down Azure resources, disable domain auto-renewal (certificate expires 2 Feb 2027) | Olga |

**Resolved since the last version**
- Should a missing database file be an error? Yes, fixed on 1 Nov.
- Session handling for login/logout: implemented on 21 Oct.
- Linting: done (golangci-lint, 30 Sep).
- Backups: automated (7 Oct and 26 Oct), but see item 7.
- Contributions from Jihad: now documented for the Azure VM notes, workflow strategy, linting, CD and backups. Still worth asking him for the reasoning behind the CD design.