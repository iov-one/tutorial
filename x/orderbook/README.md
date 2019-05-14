# Orderbook module 
---
## Requirements

This module defines the required components for orderbook. Orderbook works by matching the ask and bid orders. 

### State
- #### Order
  - ID
  - Amount
  - Price
  - Side (Ask or Bid)
- #### Order book
  - AskOrders
  - BidOrders
- #### Trade
  - MakerId
  - TakerId
  - Amount
  - Price

### Messages 
 - #### Post order
 - #### Cancel order

### Matching engine logic
---
First orders are sorted in ascending or descending order. Ask orders are sorted descendently so that the element with the highest index in the array has the lowest price and buy orders are sorted in ascending order so that the last element of the array has the highest price. When an order is posted it is inserted into one of the ask order or bid order arrays.  

#### Strategies
- ##### Best price offer strategy
  - When a matching order by price and amount is found order is placed instantly.
- ##### Partially filled order
  - If an order cannot be fulfilled entirely in one transaction, the remaining lots become a “resting order” which is included in the order book
- ##### No match
  - Recieved order becomes an resting order for future trades.
- ##### Multiple orders with same price
  - There are multiple ways to achieve this. This section is open for discussion