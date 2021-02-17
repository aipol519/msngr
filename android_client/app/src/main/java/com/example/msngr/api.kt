package com.example.msngr

import com.tinder.scarlet.Stream
import com.tinder.scarlet.WebSocket
import com.tinder.scarlet.ws.Receive
import com.tinder.scarlet.ws.Send

import io.reactivex.Flowable

interface ClientService {
    @Send
    fun sendString(message: String)
    @Receive
    fun observeText(): Stream<String>
    @Receive
    fun observeWebSocketEvent(): Flowable<WebSocket.Event>
}