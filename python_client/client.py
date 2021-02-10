import websocket
import socket
import asyncio
import hashlib
import json
import os
import sqlite3
import datetime
import aioconsole

from clientapi import requests as req

port = "80"
local_ip = ""
ws = websocket.WebSocketApp("")

client_info = req.Client("", "", "")

def receiveMessage(params):
    conn = sqlite3.connect("message.db")
    cursor = conn.cursor()

    cursor.execute("INSERT INTO messages(sender_id, receiver_id, message, time) VALUES (?, ?, ?, ?)", (params["SenderId"], params["ReceiverId"], params["Message"], datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")))

    conn.commit()
    conn.close()

funcs = {
    "receiveMessage": receiveMessage
    }

def connect(ip, port, enable_trace, on_error, on_close):
    websocket.enableTrace(enable_trace)
    global ws
    ws = websocket.WebSocketApp("ws://" + ip + ":" + port + "/",
                              on_error = on_error,
                              on_close = on_close)

    return ws

def menu():
    print("1. Register")
    print("2. Login")
    print("3. Exit")

    selector = input()
    
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
    await asyncio.sleep(5)
    
    return {
        '1' : send_message,
        '2' : view_message,
        '3' : exit,
        }[selector]

def exit():
    ws.close()
    os._exit(1)

def hash_password(password):
    return hashlib.sha224(password.encode("utf-8")).hexdigest()

def register():
    print("Enter login:")
    login = input()
    print("Enter password:")
    password = input()
    password = hash_password(password)
    print("Enter name:")
    name = input()

    global client_info
    client_info = req.Client(login, password, name)

    global ws
    ws = connect(local_ip, "80", False, on_error, on_close)
    ws.on_message = on_authorize
    ws.on_open = on_register_open
    ws.run_forever()

def on_register_open(ws):
    ws.send(req.ClientInfoRequest("registerClient", client_info).toJSON())

def on_authorize(ws, message):
    if message == "error":
        ws.close()
        menu().__call__()
    else:
        ws.on_message = on_message

        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        func = loop.run_until_complete(message_menu())
        func.__call__()

def login():
    print("Enter login:")
    login = input()
    print("Enter password:")
    password = input()
    password = hash_password(password)

    global client_info
    client_info = req.Client(login, password, "")

    global ws
    ws = connect(local_ip, "80", False, on_error, on_close)
    ws.on_message = on_authorize
    ws.on_open = on_login_open
    ws.run_forever()

def on_login_open(ws):
    ws.send(req.ClientInfoRequest("authorizeClient", client_info).toJSON())

def send_message():
    global ws

    print("Enter receiver id: ")
    receiverid = input()
    print("Enter message: ")
    message = input()

    ws.send(req.MessageRequest("sendMessage", client_info.Login, receiverid, message).toJSON())

    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    func = loop.run_until_complete(message_menu())
    func.__call__()

def view_message():
    print("Enter user id: ")
    userid = input()

    conn = sqlite3.connect("message.db")
    cursor = conn.cursor()

    cursor.execute("SELECT sender_id, receiver_id, message, time FROM messages WHERE sender_id=? AND receiver_id=?", (userid, client_info.Login))
    for message in cursor.fetchall():
        print(message)

    conn.commit()
    conn.close()
    
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    func = loop.run_until_complete(message_menu())
    func.__call__()

def on_message(ws, message):
    json_message = json.loads(message)

    if json_message["Type"] in funcs:
        funcs[json_message["Type"]](json_message)

def on_error(ws, error):
    print(error)

def on_close(ws):
    print("### closed ###")

if __name__ == "__main__":
    hostname = socket.gethostname()
    local_ip = socket.gethostbyname(hostname)

    menu().__call__()
    