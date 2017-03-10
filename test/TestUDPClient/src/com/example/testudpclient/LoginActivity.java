package com.example.testudpclient;

import java.io.InputStream;

import org.json.JSONObject;

import android.app.Activity;
import android.content.Intent;
import android.os.Bundle;
import android.view.View;
import android.view.View.OnClickListener;
import android.widget.Button;
import android.widget.EditText;

import com.giant.sdk.net.GWebClient;
import com.giant.sdk.log.GLog;

public class LoginActivity extends Activity {
	Button btn_login;
	EditText txt_account;
	EditText txt_password;
	EditText txt_ip;
	
	@Override
	protected void onCreate(Bundle savedInstanceState) {
		super.onCreate(savedInstanceState);
		setContentView(R.layout.activity_login);
		
		btn_login = (Button)findViewById(R.id.btn_login);
		txt_account = (EditText)findViewById(R.id.txt_account);
		txt_password = (EditText)findViewById(R.id.txt_password);
		txt_ip = (EditText)findViewById(R.id.txt_ip);
		
		btn_login.setOnClickListener(new OnClickListener(){
			@Override
			public void onClick(View view) {
				//GIMDemo.Instance().login(txt_account.getText().toString(), txt_password.getText().toString());
				Thread thread = new Thread(new Runnable(){
					public void run() {
						try{
							String url = "http://" + txt_ip.getText().toString() + "/login?useraccount=" + txt_account.getText().toString();
							GWebClient webclient = new GWebClient(url);
							webclient.openConnection();
							
							InputStream is = webclient.openStream();
							
							String str = null;
							String result = "";
							while((str = webclient.readLine()) != null){
								result += str;
							}
							webclient.close();
							webclient = null;
							GLog.d(result);
							
							JSONObject roomJson = new JSONObject(result);
							
							if(!roomJson.has("error")){
								//go to chatroomlist
								GlobalInfo.sessionId = roomJson.getLong("uid");
								GlobalInfo.ip = txt_ip.getText().toString();
								
								/* 新建一个Intent对象 */
				                Intent intent = new Intent();
				                //intent.putExtra("name","LeiPei");
				                /* 指定intent要启动的类 */
				                intent.setClass(LoginActivity.this, ChatRoomListActivity.class);
				                /* 启动一个新的Activity */
				                LoginActivity.this.startActivity(intent);
				                /* 关闭当前的Activity */
				                LoginActivity.this.finish();
							}else{
								GLog.e(result);
							}
						}catch(Exception e){
							e.printStackTrace();
						}
						//go to chatroomlistactivity
					}
				});
				thread.start();
			}
		});
	}
}
