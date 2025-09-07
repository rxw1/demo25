Services
--------

- Created/Placed: the customer has selected products and submitted the order.
- Confirmed/Accepted: the system (and possibly payment provider) has validated
  it, stock checked, etc.
- Processed: the order is being prepared: items are allocated, picked, packed.
- Shipped/Dispatched: handed over to carrier.
- Delivered: reached the customer.
- Closed/Completed: everything went fine, transaction settled.
- Cancelled/Returned: side branches in case it doesn’t.

- Catalog service (product info, prices, availability metadata)
- Inventory service (actual stock counts, reservations)
- Order service (lifecycle, status, line items)
- Payment service (authorization, capture, refund)
- Shipping / Fulfillment service (address validation, shipping labels, tracking)
- Customer service (accounts, addresses, history)
