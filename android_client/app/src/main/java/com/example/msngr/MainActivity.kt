package com.example.msngr

import android.annotation.SuppressLint
import androidx.appcompat.app.AppCompatActivity
import android.os.Bundle
import android.util.Log
import com.google.gson.Gson

import com.tinder.scarlet.Scarlet
import com.tinder.scarlet.WebSocket
import com.tinder.scarlet.streamadapter.rxjava2.RxJava2StreamAdapterFactory
import com.tinder.scarlet.websocket.okhttp.newWebSocketFactory
import io.reactivex.schedulers.Schedulers
import okhttp3.OkHttpClient
import java.net.InetAddress

const val NODE_SERVER_URL = "ws://192.168.81.226:80"

class MainActivity : AppCompatActivity() {

    @SuppressLint("CheckResult")
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        val okHttpClient = OkHttpClient.Builder()
                .readTimeout(0, java.util.concurrent.TimeUnit.MILLISECONDS)
                .build()

        val client = Scarlet.Builder()
                .webSocketFactory(okHttpClient.newWebSocketFactory(NODE_SERVER_URL))
                .addStreamAdapterFactory(RxJava2StreamAdapterFactory())
                .build()
                .create<ClientService>()

        var gson = Gson()
        val testString = gson.toJson(TestRequest("testRequest", "testData"))

        client.observeWebSocketEvent()
            .filter { it is WebSocket.Event.OnConnectionOpened<*> }
            .subscribe({
                client.sendString("""{"TestRequestData":"testData","Type":"testRequest"}""")
            }, { e ->
                print(e)
            })
    }
}