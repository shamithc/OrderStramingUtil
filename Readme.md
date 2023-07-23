Define sorted set with 
  key: bids-{market-name}
  values: bids-{market-name}-{price}
  score: price

Define Hash with
  key: bids-sum
  field: {market-name}-price
  value: quantity




curl http://localhost:8080/updateOrderBook \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{"order_type": "buy", "market": "BTC-INR", "price": 100.0, "quantity": 5}'



"{ "user_id": "user_uuid", "order_type": "BID/ASK", "order_execution_type": "LIMIT/MARKET", "fill_or_kill": true/false, "price": u64, "amount": u64, "pair": "pair_uuid" }"


buy-BTC-INR


curl http://localhost:8080/fetchOrderBook/BTC-INR