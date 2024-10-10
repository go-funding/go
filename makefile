include .env

run_google_chrome:
	/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222 --kiosk --disable-web-security --user-data-dir=./@data

generate_source_maps:
	sourcemapper -output test -url ./output/1728008365255717000/lt.smebank.ebank/React/build/assets/index-C06-YtBS.js.map -output ./output/1728008365255717000/lt.smebank.ebank/React/build/assets/index-C06-YtBS/
