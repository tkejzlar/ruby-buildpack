require 'webrick'
require 'typhoeus'

server = WEBrick::HTTPServer.new(Port: ENV.fetch('PORT', 8080))
server.mount_proc '/' do |req, res|
  res.body = Typhoeus.get("www.example.com", followlocation: true).body
end

trap('INT') { server.stop }
server.start
