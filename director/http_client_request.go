package director

func (d DirectorImpl) NewHTTPClientRequest() ClientRequest {
	return d.client.clientRequest
}
