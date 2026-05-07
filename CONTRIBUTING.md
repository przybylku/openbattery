# Contributing to openbattery

Dzięki, że chcesz pomóc! 🎉

## Jak zgłaszać błędy

1. Sprawdź, czy ktoś już nie zgłosił tego problemu w [Issues](../../issues).
2. Jeśli nie – otwórz nowy Issue i użyj szablonu **Bug report**.
3. Podaj jak najwięcej szczegółów:
   - wersja macOS,
   - wersja `openbattery` (lub commit),
   - co się stało i co powinno się stać,
   - jak to powtórzyć,
   - zrzuty ekranu (jeśli dotyczy UI).

## Jak proponować nowe funkcje

1. Otwórz Issue z szablonem **Feature request**.
2. Opisz:
   - co chcesz osiągnąć,
   - dlaczego to potrzebne,
   - ew. pomysł na implementację.

## Jak wprowadzać zmiany (Pull Request)

1. **Fork** repo i sklonuj lokalnie.
2. Upewnij się, że masz zainstalowane:
   - [Go](https://go.dev/dl/) ≥ 1.21
   - `golangci-lint` (opcjonalnie, ale przydatne)
3. Uruchom testy i lintery przed commitem:
   ```bash
   go vet ./...
   go test -race -v ./...
   golangci-lint run
   ```
4. Stwórz branch z opisową nazwą:
   ```bash
   git checkout -b fix/short-description
   # lub
   git checkout -b feat/short-description
   ```
5. Wprowadź zmiany. Staraj się:
   - trzymać istniejący styl kodu,
   - dodać testy jeśli to możliwe,
   - nie zostawiać zbędnych `// TODO` bez wyjaśnienia.
6. Commituj z opisowymi wiadomościami (po angielsku lub polsku – byle z sensem):
   ```
   fix: correct percentage rounding in dashboard
   feat: add dark mode toggle
   ```
7. Otwórz Pull Request i wypełnij szablon.

## Jak uruchomić lokalnie

```bash
go mod download
go build -o openbattery .
./openbattery
```

## Testy

Projekt używa standardowego `go test`:

```bash
go test ./...
```

Jeśli dodajesz nowy moduł, postaraj się napisać do niego testy.

## Style guide

- Formatowanie: `gofmt` (lub `goimports`).
- Nazwy po angielsku.
- Komentarze eksportowanych funkcji/typów zgodnie z konwencją Go.
- Nie commituj binarki `openbattery` – jest w `.gitignore`.

## Pytania?

Otwórz [Discussion](../../discussions) lub napisz w Issue – chętnie pomogę.
