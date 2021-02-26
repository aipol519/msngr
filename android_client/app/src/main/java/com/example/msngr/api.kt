package com.example.msngr

import com.tinder.scarlet.WebSocket
import com.tinder.scarlet.ws.Receive
import com.tinder.scarlet.ws.Send

import io.reactivex.Flowable
import java.security.MessageDigest

interface ClientService {
    @Send
    fun sendString(message: String)
    @Send
    fun sendBytes(message: ByteArray)
    @Receive
    fun observeBytes(): Flowable<ByteArray>
    @Receive
    fun observeWebSocketEvent(): Flowable<WebSocket.Event>
}

fun hashPassword(Password: String): String {
    val messageDigest = MessageDigest.getInstance("SHA-224")
    messageDigest.update(Password.toByteArray())

    val digest = messageDigest.digest()
    val buffer = StringBuffer()

    for (i in digest.indices) {
        var hex_string = Integer.toHexString(0xFF and digest[i].toInt())
        if (hex_string.length == 1) hex_string = "0$hex_string"
        buffer.append(hex_string)
    }

    return buffer.toString()
}

open class Request(val Type: String) {}

class TestRequest(Type: String, val TestRequestData: String) : Request(Type) {}
class InfoRequest(Type: String, val Info: String) : Request(Type) {}

class AuthorizeClientRequest(Type: String, val Login: String, val PasswordHash: String) : Request(Type) {}