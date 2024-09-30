## Notes
- Clients send commands to the Redis server as RESP arrays.
- that's y PING start with asterisk *
- e.g. : *1\r\n$4\r\nPING\r\n
- Array format: `*<number-of-elements>\r\n<element-1>...<element-n>`
- Bulk string format: `$<length>\r\n<data>\r\n`
