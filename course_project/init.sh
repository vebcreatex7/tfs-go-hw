curl -v -X POST -H "Content-Type: application/json" --data '{"symbol":"pi_xbtusd"}' localhost:8000/bot/set_symbol
curl -v -X POST -H "Content-Type: application/json" --data '{"period":"1m"}' localhost:8000/bot/set_period
