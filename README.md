## shinsei (新生)
> Shinsei is a (Proxy Server with Load Balancing) project created in Go, and means "rebirth" or "new life" (新生).

This is a simple proxy server with load balancing capability implemented in Go. It uses the Round Robin algorithm to evenly distribute requests among available backend servers.

## Features

- **Round-robin Load Balancing:** The proxy server evenly distributes requests among the listed destination servers.

- **Simultaneous Connections Limit:** A limit on simultaneous connections has been implemented to prevent overload on the proxy server and destination servers. This is achieved using a semaphore to control the maximum number of active connections.

- **HTTP and TCP Proxy:** The proxy server supports both HTTP and TCP connections, allowing for forwarding different types of traffic.

- **Error Logging:** Errors encountered during connection to destination servers or during request forwarding are logged for debugging and monitoring purposes.
