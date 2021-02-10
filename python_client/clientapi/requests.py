import json

class Client:
    Login = ""
    PasswordHash = ""
    Name = ""

    def __init__(self, login, passwordhash, name):
        self.Login = login
        self.PasswordHash = passwordhash
        self.Name = name

class Request:
    Type = ""

    def __init__(self, type):
        self.Type = type

    def toJSON(self):
        return json.dumps(self, default = lambda o: o.__dict__, sort_keys = True, indent = 4)

class ClientInfoRequest(Request):
    Client = Client("", "", "")

    def __init__(self, type, client):
        self.Type = type
        self.Client = client

class MessageRequest(Request):
    SenderId = ""
    ReceiverId = ""
    Message = ""
    
    def __init__(self, type, senderid, receiverid, message):
        self.Type = type
        self.SenderId = senderid
        self.ReceiverId = receiverid
        self.Message = message