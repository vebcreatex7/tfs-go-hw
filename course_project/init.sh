curl -v -X POST -H "Content-Type: application/json" --data '{"symbol":"pi_xbtusd"}' localhost:8000/bot/set_symbol
curl -v -X POST -H "Content-Type: application/json" --data '{"period":"5m"}' localhost:8000/bot/set_period
