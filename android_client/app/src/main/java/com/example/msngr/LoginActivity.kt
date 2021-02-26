package com.example.msngr

import android.annotation.SuppressLint
import androidx.appcompat.app.AppCompatActivity
import android.os.Bundle
import android.util.Log
import android.widget.Button
import android.widget.EditText
import com.google.gson.Gson

import com.tinder.scarlet.Scarlet
import com.tinder.scarlet.streamadapter.rxjava2.RxJava2StreamAdapterFactory
import com.tinder.scarlet.websocket.okhttp.newWebSocketFactory
import okhttp3.OkHttpClient
import java.security.MessageDigest

const val NODE_SERVER_URL = "ws://192.168.50.201:80"

class MainActivity : AppCompatActivity() {

    private lateinit var loginEditText: EditText
    private lateinit var passwordEditText: EditText

    private lateinit var loginButton: Button

    @SuppressLint("CheckResult")
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_login)

        val okHttpClient = OkHttpClient.Builder()
                .readTimeout(0, java.util.concurrent.TimeUnit.MILLISECONDS)
                .build()

        val client = Scarlet.Builder()
                .webSocketFactory(okHttpClient.newWebSocketFactory(NODE_SERVER_URL))
                .addStreamAdapterFactory(RxJava2StreamAdapterFactory())
                .build()
                .create<ClientService>()

        val gson = Gson()

        loginEditText = findViewById(R.id.loginEditText)
        passwordEditText = findViewById(R.id.passwordEditText)

        loginButton = findViewById(R.id.loginButton)
        loginButton.setOnClickListener {
            val testString = gson.toJson(AuthorizeClientRequest("authorizeClient", loginEditText.text.toString(), hashPassword(passwordEditText.text.toString())))
            client.sendBytes(testString.toByteArray())
        }

        client.observeBytes()
                .subscribe({
                    message -> Log.d("testMesssageReceived", message.toString(Charsets.UTF_8))
                }, { e ->
                    print(e)
                })
    }
}