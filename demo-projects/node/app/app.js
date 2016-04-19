var http = require('http');

http.createServer(function (req, res) {
  res.writeHead(200, {'Content-Type': 'text/plain'});
  var ip = req.connection.remoteAddress;
  console.log('access from IP ' + ip);
  res.end('Hello ' + ip);
}).listen(8080, '0.0.0.0');