#!/bin/sh

# CREATE TABLE orderbook (timestamp DATETIME, code TEXT, stock INTEGER, sell_price INTEGER, buy_price INTEGER, PRIMARY KEY (timestamp, code));
# WITH X AS (SELECT code, MAX(timestamp) as tx FROM orderbook GROUP BY code) SELECT O.* FROM orderbook O JOIN X on O.timestamp = X.tx;

# CREATE TABLE transactions (timestamp DATETIME DEFAULT CURRENT_TIMESTAMP, code TEXT, quantity INTEGER, price INTEGER, side TEXT, PRIMARY KEY (timestamp, code));
# CREATE TABLE shadow_transactions (timestamp DATETIME DEFAULT CURRENT_TIMESTAMP, code TEXT, quantity INTEGER, price INTEGER, side TEXT, PRIMARY KEY (timestamp, code));
# CREATE TABLE market_parameters (enabled BOOLEAN, code TEXT, theo INTEGER, max_stock INTEGER, min_stock INTEGER, PRIMARY KEY (code));

INSERT INTO market_parameters (enabled, code, theo, max_stock, min_stock) VALUES 
    (true,  "wolf_hair",       210,    200,    0),
    (true,  "egg",             4,      1000,   100),
    (true,  "jasper_crystal",  3200,   50,     3),
    (false, "milk_bucket",     35,     25,     0),
    (false, "ogre_skin",       435,    50,     0),
    (true,  "feater",          45,     250,    0);