# tailgate

Tailgate hooks server log files to Slack channels.

## OPTIONS

  **-path `<path>`**
  
  Path of the log file to read.
  
  **-match `<string>`**
  
  Send message to Slack channel when log line contains <string>.
  
  **-httpserver `<bool>`**
  
  Run tailgate's inbuilt http server which serves the file you're tailing to http://localhost:8080/ (last lines first).
