import asyncio
import websockets
import os
import socket
import hashlib
import sqlite3
import datetime
import json
import aioconsole

from clientapi import requests as req
from tabulate import tabulate

port = "80"
local_ip = ""
handler = None
client_info = None

class Handler:
    def __init__(self):
        self.ws = None
        self.loop = asyncio.get_event_loop()
        self.loop.run_until_complete(self.__async__connect())

    async def __async__connect(self):
        uri = "ws://" + local_ip + ':' + port
        self.ws = await websockets.connect(uri)

    async def listen(self):
        while True:
            message = await self.ws.recv()
            json_message = json.loads(message)

            loop = asyncio.get_event_loop()

            if json_message["Type"] in funcs:
                loop.create_task(funcs[json_message["Type"]](json_message))
        
    async def register(self, client_info):
        await self.ws.send(req.ClientInfoRequest("registerClient", client_info).toJSON())

    async def login(self, client_info):
        await self.ws.send(req.ClientInfoRequest("authorizeClient", client_info).toJSON())

    async def send_message(self, client_info):
        print("Enter receiver id: ")
        receiverid = await aioconsole.ainput()
        print("Enter message: ")
        message = await aioconsole.ainput()

        await self.ws.send(req.MessageRequest("sendMessage", client_info.Login, receiverid, message).toJSON())

        loop = asyncio.get_event_loop()

        mm_loop = loop.create_task(message_menu())
        func = await mm_loop

        loop.create_task(func())

def hash_password(password):
    return hashlib.sha224(password.encode("utf-8")).hexdigest()

async def receiveMessage(params):
    print("Message received from:", params["SenderId"], ':', params["Message"])
    
    conn = sqlite3.connect("message.db")
    cursor = conn.cursor()

    cursor.execute("INSERT INTO messages(sender_id, receiver_id, message, time) VALUES (?, ?, ?, ?)", (params["SenderId"], params["ReceiverId"], params["Message"], datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")))

    conn.commit()
    conn.close()

async def infoRequest(params):
    loop = asyncio.get_event_loop()

    if params["Info"] == "success":
        mm_loop = loop.create_task(message_menu())
        func = await mm_loop
    elif params["Info"] == "error":
        m_loop = loop.create_task(menu())
        func = await m_loop

    loop.create_task(func())

funcs = {
    "receiveMessage": receiveMessage,
    "infoRequest": infoRequest,
    }

async def menu():
    print("1. Register")
    print("2. Login")
    print("3. Exit")

    selector = await aioconsole.ainput()
    
    return {
        '1' : register,
        '2' : login,
        '3' : exit,
        }[selector]

async def message_menu():
    print("1. Send message")
    print("2. View messages")
    print("3. Exit")

    selector = await aioconsole.ainput()
    
    return {
        '1' : send_message,
        '2' : view_message,
        '3' : exit,
        }[selector]

async def register():
    print("Enter login:")
    login = await aioconsole.ainput()
    print("Enter password:")
    password = await aioconsole.ainput()
    password = hash_password(password)
    print("Enter name:")
    name = await aioconsole.ainput()

    global client_info
    client_info = req.Client(login, password, name)

    loop = asyncio.get_event_loop()
    loop.create_task(handler.register(client_info))
    
    return

async def login():
    print("Enter login:")
    login = await aioconsole.ainput()
    print("Enter password:")
    password = await aioconsole.ainput()
    password = hash_password(password)

    global client_info
    client_info = req.Client(login, password, "")

    loop = asyncio.get_event_loop()
    loop.create_task(handler.login(client_info))
    
    return

async def exit():
    handler.ws.close()
    os._exit(1)

async def send_message():
    loop = asyncio.get_event_loop()
    loop.create_task(handler.send_message(client_info))
   
async def view_message():
    print("Enter user id: ")
    userid = await aioconsole.ainput()

    conn = sqlite3.connect("message.db")
    cursor = conn.cursor()

    cursor.execute("SELECT sender_id, receiver_id, message, time FROM messages WHERE sender_id=? AND receiver_id=?", (userid, client_info.Login))

    messages = cursor.fetchall()

    print(tabulate(messages, headers=["Sender", "Receiver", "Message", "Datetime"]))

    conn.commit()
    conn.close()
    
    loop = asyncio.get_event_loop()

    mm_loop = loop.create_task(message_menu())
    func = await mm_loop

    loop.create_task(func())

hostname = socket.gethostname()
local_ip = socket.gethostbyname(hostname)

loop = asyncio.new_event_loop()
asyncio.set_event_loop(loop)
func = loop.run_until_complete(menu())

handler = Handler()

loop.create_task(handler.listen())
loop.create_task(func())
loop.run_forever()