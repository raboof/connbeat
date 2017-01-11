from BaseHTTPServer import BaseHTTPRequestHandler,HTTPServer
import simplejson

PORT = 7070


class connHandler(BaseHTTPRequestHandler):
	state = None

	def log_message(self, format, *args):
		# print("nope")
		42

	def do_POST(self):
		self.state['n_posts'] = self.state['n_posts'] + 1
		# print("got POST %d" % self.state['n_posts'])
		data_string = self.rfile.read(int(self.headers['Content-Length']))
		data = simplejson.loads(data_string)
		# print(data)
		if 'container' in data.keys():
			we = data['container']['id'][0:10]
		else:
			we = data['process']

		if 'local_ip' in data.keys():
			local_pair = "{}:{}".format(data['local_ip'], data['local_port'])
			self.state['local_identities'][local_pair] = we
		else:
			local_pair = '*:{}'.format(data['local_port'])
		for local_ip in data['container']['local_ips']:
			self.state['local_identities']["{}:{}".format(local_ip, data['local_port'])] = we

		if 'remote_port' in data.keys():
			remote_pair = "{}:{}".format(data['remote_ip'], data['remote_port'])
			if remote_pair in self.state['local_identities'].keys():
				print("{} (on {}) is connected to {} (on {})".format(
					local_pair,
					we,
					remote_pair,
					self.state['local_identities'][remote_pair]
				))

		self.send_response(200)
		self.end_headers()

connHandler.state = {}
connHandler.state['n_posts'] = 0
connHandler.state['local_identities'] = {}
httpd = HTTPServer(("", PORT), connHandler)

print "serving at port", PORT
httpd.serve_forever()
