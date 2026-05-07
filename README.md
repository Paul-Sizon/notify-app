# Signal Monitor

**Using agents to improve signal-to-noise ratio.**

Cross-platform client (**iOS, Android, Desktop, Web**) backed by a Go agent server.

Tell it what you care about. It watches the open web on a schedule, extracts only matching signals, dedupes, and pushes a notification when something new appears.

> Built with Kotlin Multiplatform (Compose for Android/Desktop, SwiftUI for iOS, React + Kotlin/JS for Web) and a single Go backend.

## Demo



https://github.com/user-attachments/assets/05757784-e35f-4056-ba32-4624f31471ec



> To embed an inline player on GitHub, drag-drop `ios-demo.mp4` into a GitHub issue or PR comment, copy the resulting `https://github.com/user-attachments/...` URL, then replace the line above with `<video src="<that-url>" controls></video>`. GitHub only renders `<video>` from its own asset host, not from repo-relative paths.

## Why

> **Q:** Can't I use X, Instagram, TikTok, any FYP page, or email to get news?
>
> **A:** Yes, but those feeds are noisy, and they won't surface narrow topics unless you actively dig.

### Example: specialized news for an international SaaS

- small changes in crypto regulations in Vietnam
- Visa/Mastercard policy updates
- new competitor app on the local market

Stuff that may not get to your FYP page.

## How it works

1. Create a **subscription**: natural-language query, cadence (≥5 min), and type (`event` or `news`).
2. Server agent runs on schedule:
   - Brave Web Search pulls top results (title, URL, snippet, page_age).
   - OpenAI structured extraction filters to matching signals only.
   - Fingerprint dedup vs. prior signals for that subscription.
3. New signals trigger a push notification to your device.

<img width="301" height="655" alt="q01-coldplay-detail" src="https://github.com/user-attachments/assets/533a34cc-4aec-48ee-8bd5-3df23bdd6cf9" />


## Repo layout

| Path | What |
|---|---|
| `server/` | Go backend: agent, API, scheduler, APNs. See [`server/README.md`](./server/README.md). |
| `composeApp/` | Kotlin Multiplatform UI: Android + Desktop (JVM) from shared Compose source. |
| `iosApp/` | SwiftUI host for iOS, calls into `shared`. |
| `webApp/` | React + Kotlin/JS web client. |
| `shared/` | KMP shared module (models, networking). |
| `server/migrations/` | Goose SQL migrations. |

## Quickstart

### Server

```bash
cd server
make up && make migrate
make run
```

Required env: `DATABASE_URL`, `OPENAI_API_KEY`, `BRAVE_SEARCH_API_KEY`. APNs vars optional; without them the server logs notifications to stdout.

### Clients

| Target | Command |
|---|---|
| iOS | open `iosApp/` in Xcode, run |
| Android | `./gradlew :composeApp:assembleDebug` |
| Desktop | `./gradlew :composeApp:run` |
| Web | `./gradlew :shared:jsBrowserDevelopmentLibraryDistribution && npm install && npm run start` |

## API

See [`server/README.md`](./server/README.md#api) for endpoint reference.
