#!/bin/sh

# CREATE TABLE orderbook (timestamp DATETIME, code TEXT, stock INTEGER, sell_price INTEGER, buy_price INTEGER, PRIMARY KEY (timestamp, code));
# WITH X AS (SELECT code, MAX(timestamp) as tx FROM orderbook GROUP BY code) SELECT O.* FROM orderbook O JOIN X on O.timestamp = X.tx;