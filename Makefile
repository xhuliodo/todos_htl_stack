export PORT=:3000

run:
	templ generate
	npx tailwindcss --minify -o assets/tailwind.css
	PORT=$(PORT) go run main.go