#!/bin/sh

# CREATE TABLE orderbook (timestamp DATETIME, code TEXT, stock INTEGER, sell_price INTEGER, buy_price INTEGER, PRIMARY KEY (timestamp, code));
# WITH X AS (SELECT code, MAX(timestamp) as tx FROM orderbook GROUP BY code) SELECT O.* FROM orderbook O JOIN X on O.timestamp = X.tx;

# CREATE TABLE transactions (timestamp DATETIME, code TEXT, quantity INTEGER, price INTEGER, side TEXT, PRIMARY KEY (timestamp, code));
# CREATE TABLE market_parameters (code TEXT, theo INTEGER, max_stock INTEGER. min_stock INTEGER, PRIMARY KEY (code));