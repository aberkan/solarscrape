# Solar usage stats (Locus Energy)

Fetches daily solar production (Wh) for a given month or year from the Locus Energy API and prints `Date` and `Wh_sum` for piping into other apps.

## Build

From the project directory:

```bash
go build ./...
```

This produces `solarscrape` on Unix/macOS and `solarscrape.exe` on Windows.

## Usage

```bash
solarscrape -token <bearer-token> (-month YYYY-MM | -year YYYY)
```

On Windows use `solarscrape.exe` instead of `solarscrape`.

Examples:

```bash
solarscrape -token your-token-here -month 2026-01
solarscrape -token your-token-here -year 2026
```

Output (tab-separated, one line per day):

```
2026-01-25	0.000
2026-01-26	380.000
2026-01-27	9100.000
...
```

- **Month** (`-month`): `YYYY-MM` (e.g. `2026-01`). Data from the first day of that month up to **today** (in America/Edmonton), or the end of the month if today is later.
- **Year** (`-year`): `YYYY` (e.g. `2026`). Use instead of `-month` to get the full year: Jan 1 up to today or Dec 31, whichever is earlier.
- **Token**: Your API bearer token. Do not commit it; pass it on the command line or via env.

## How to get the bearer token from the browser

1. Open [https://locusnoc.datareadings.com](https://locusnoc.datareadings.com) and log in.
2. Open Developer Tools:
   - **Chrome / Edge**: `F12` or `Ctrl+Shift+J` (Windows/Linux), `Cmd+Option+J` (Mac).
   - **Firefox**: `F12` or `Ctrl+Shift+K` (Windows/Linux), `Cmd+Option+K` (Mac).
3. Go to the **Network** tab.
4. In the page, select a month so that the site loads that month's data.
5. In the Network list, find a request to `api.locusenergy.com` (e.g. URL containing `/sites/.../data?...`). Click it.
6. In the request **Headers**, look for **Request Headers** and find **Authorization**.  
   The value will look like: `Bearer abc123def456...`.  
   The part after `Bearer ` is your token.
7. Use that value with the `-token` flag.  
   **Security**: Treat this like a password. Don't share it or commit it to version control.

## Timezone

Data is requested with timezone `America/Edmonton`. "Today" for the end date is evaluated in that timezone.
