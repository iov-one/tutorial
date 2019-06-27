# Orderbook module 
---
## Requirements

This module defines the required components for orderbook. Orderbook works by matching the ask and bid orders. 

### State
- #### Order state
  - Open: *means order is still pending*
  - Done: *means order has been executed*
  - Cancel: *means order has been cancelled*
- #### Side
  - Ask: *defines trader wants to buy*
  - Bid *defines trader wants to sell*
- #### Order
  - ID
  - Trader: *identity of trader*
  - OrderBookID: *orderbook id of this order belongs to*
  - Side: *side of this order*
  - OrderState: *defines if order is done, cancelled, or open*
  - OriginalOffer: *amount of offer at the time of creation*
  - RemainingOffer: *amount of pending trade*
  - Price: *price of offer*
  - TradeIDs: *trades that have been executed*
  - CreatedAt: *creation time of offer*
  - UpdatedAt: *update time of offer. Updated whenever order state changes*
- #### Trade
  - ID
  - OrderBookID: *ID of the orderbook trade happened at*
  - OrderID: *ID of order this trade is originated*
  - Taker: *identity of taker that accepted to buy*
  - Maker: *identity of maker that accepted to sell*
  - MakerPaid: *amount maker paid to settle the trade*
  - TakerPaid: *amount taker paid to settle the trade*
  - ExecutedAt: *defines the trade execution time*
- #### Order book
  - ID
  - MarketID: *market this orderbook belongs to*
  - AskTicker: *Ticker of ask side*
  - BidTicker: *Ticker of bid side*
  - TotalAskCount: *number of available ask orders*
  - TotalBidCount: *number of available bid orders*
- #### Market
  - ID
  - Owner: *identity of owner of this market*
  - Name: *name of the market*

### Messages 
 - #### Post order
    - MarketID: *market that order is posted to*
    - AskTicker: *Ticker of ask side*
    - BidTicker: *Ticker of bid side*
 - #### Cancel order
    - OrderID: *Order that wanted to be cancelled*

### Order and Trade relation
Trade is full/partial offer that happened between traders

### Market and Orderbook relation
The blockchain may have multiple markets. Each market may have multiple orderbooks. Each token pair can only have one orderbook per market.
There is no global chain owner, but each market has one that adds orderbooks and sets fees. This could be one person, a multisig, or a governance contract (Dao)

When adding an orderbook, market with given market id is checked, then look up the owner *for that market*. Rather than having one owner that can create orderbooks for all markets, each market stores who can update it.

### Matching engine logic
---
First orders are sorted in ascending or descending order. Ask orders are sorted descendently so that the element with the highest index in the array has the lowest price and buy orders are sorted in ascending order so that the last element of the array has the highest price. When an order is posted it is inserted into one of the ask order or bid order arrays.  

#### Strategies
- ##### Best price offer strategy
  - When a matching order by price and amount is found order is placed instantly.
- ##### Partially filled order
  - If an order cannot be fulfilled entirely in one transaction, the remaining lots become a “resting order” which is included in the order book. Resting orders are prioritised and fulfilled when matching orders are received.
- ##### No match
  - Recieved order becomes an resting order for future trades.
- ##### Multiple orders with same price
  - There are multiple ways to achieve this. This section is open for discussion
