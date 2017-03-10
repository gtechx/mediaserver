package com.example.testudpclient;

import java.io.InputStream;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;

import org.json.JSONArray;
import org.json.JSONObject;

import com.giant.sdk.log.GLog;
import com.giant.sdk.net.GWebClient;

import android.app.Activity;
import android.content.Context;
import android.content.Intent;
import android.os.Bundle;
import android.view.LayoutInflater;
import android.view.View;
import android.view.View.OnClickListener;
import android.widget.AdapterView;
import android.widget.AdapterView.OnItemClickListener;
import android.widget.Button;
import android.widget.LinearLayout.LayoutParams;
import android.widget.ListView;
import android.widget.SimpleAdapter;

public class ChatRoomListActivity extends Activity {
	ListView lv_chatroom;
	Button btn_createroom;
	Button btn_refresh;
	ChatRoomListViewAdapter adapter;

	ArrayList<RoomInfo> mRoomList = new ArrayList<RoomInfo>();
	LayoutInflater mInflater;
	
	public class Room{
		public String id;
		public String ip;
		public int port;
		public ArrayList<Room> mSubRoom;
		
		public Room(JSONObject obj){
			try{
				ip = obj.getString("ip");
				port = obj.getInt("port");
				
				if(obj.has("id")){
					id = obj.getString("id");
				}
				
				if(obj.has("subroom")){
					mSubRoom = new ArrayList<Room>();
					JSONArray subroomarray = obj.getJSONArray("subroom");
					
					for(int i = 0; i < subroomarray.length(); ++i){
						Room subroom = new Room(subroomarray.getJSONObject(i));
						mSubRoom.add(subroom);
					}
				}
			}catch(Exception e){
				e.printStackTrace();
			}
		}
	}
	
	public class RoomInfo{
		public String type;
		public boolean haspassword;
		public Room room;
		
		public RoomInfo(JSONObject obj){
			try{
				type = obj.getString("type");
				haspassword = obj.getInt("haspassword") == 1;
				JSONObject roomobj = obj.getJSONObject("room");
				room = new Room(roomobj);
			}catch(Exception e){
				e.printStackTrace();
			}
		}
	}
	
	protected void parseRoomInfo(String json){
		mRoomList.clear();
		try{
			JSONArray roomarray = new JSONArray(json);
			GLog.d("json array length:" + roomarray.length());
			for(int i = 0; i < roomarray.length(); ++i){
				JSONObject roomInfobj = roomarray.getJSONObject(i);
				RoomInfo rinfo = new RoomInfo(roomInfobj);
				mRoomList.add(rinfo);
			}
		}catch(Exception e){
			e.printStackTrace();
		}
	}
	
	@Override
	protected void onCreate(Bundle savedInstanceState) {
		super.onCreate(savedInstanceState);
		setContentView(R.layout.activity_chatroomlist);
		mInflater = LayoutInflater.from(this);
		
		lv_chatroom = (ListView)this.findViewById(R.id.lv_chatroom);
		btn_createroom = (Button)this.findViewById(R.id.btn_createroom);
		btn_refresh = (Button)this.findViewById(R.id.btn_refresh);
		
		final Context con = (Context)this;
		
		btn_createroom.setOnClickListener(new OnClickListener(){
			@Override
			public void onClick(View view) {
				View menuview = mInflater.inflate(R.layout.menu_createroom, null);
				LayoutParams params = new LayoutParams(LayoutParams.MATCH_PARENT, LayoutParams.MATCH_PARENT);
				((Activity)con).addContentView(menuview, params);
			}
		});
		
		btn_refresh.setOnClickListener(new OnClickListener(){
			@Override
			public void onClick(View view) {
				getRoomList();
			}
		});
//		lv_chatroom.setOnItemClickListener(new OnItemClickListener(){
//			@Override
//		    public void onItemClick(AdapterView<?> adapterView, View view, int position,long id) {
//				RoomInfo rinfo = mRoomList.get(position);
//				GLog.d("info:" + rinfo.type + " ip=" + rinfo.room.ip + " port=" + rinfo.room.port);
//		    }
//		});
		getRoomList();
	}
	
	protected void getRoomList(){
		final Context con = (Context)this;
		
		Thread thread = new Thread(new Runnable(){
			public void run(){
				try{
					String url = "http://" + GlobalInfo.ip + "/listrooms?sessionid=" + GlobalInfo.sessionId;
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
					
					parseRoomInfo(result);
					
					adapter = new ChatRoomListViewAdapter(con, mRoomList);
//					adapter = new SimpleAdapter(con, mData, R.layout.item_chatroom, 
//							new String[]{"tv_chatroom","tv_type"},new int[]{R.id.tv_chatroom,R.id.tv_type});
					lv_chatroom.setAdapter(adapter);
				}catch(Exception e){
					e.printStackTrace();
				}
			}
		});
		
		thread.start();
	}
	
	public void onCreateZhubo(View view){
		Thread thread = new Thread(new Runnable(){
			public void run(){
				try{
					String url = "http://" + GlobalInfo.ip + "/create?type=zhubo&sessionid=" + GlobalInfo.sessionId;
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
					
					Room room = new Room(new JSONObject(result));
					
					GlobalInfo.isZhubo = true;
					GlobalInfo.roomType = "zhubo";
					GlobalInfo.roomIp = room.ip;
					GlobalInfo.roomPort = room.port;
					GlobalInfo.subroomIp = room.mSubRoom.get(0).ip;
					GlobalInfo.subroomPort = room.mSubRoom.get(0).port;
					
					/* 新建一个Intent对象 */
		            Intent intent = new Intent();
		            //intent.putExtra("name","LeiPei");
		            /* 指定intent要启动的类 */
		            intent.setClass(ChatRoomListActivity.this, ChatRoomActivity.class);
		            /* 启动一个新的Activity */
		            ChatRoomListActivity.this.startActivity(intent);
		            /* 关闭当前的Activity */
		            ChatRoomListActivity.this.finish();
				}catch(Exception e){
					
				}
			}
		});
		thread.start();
	}
	
	public void onCreateZiyou(View view){
		Thread thread = new Thread(new Runnable(){
			public void run(){
				try{
					String url = "http://" + GlobalInfo.ip + "/create?sessionid=" + GlobalInfo.sessionId;
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
					
					Room room = new Room(new JSONObject(result));
					
					GlobalInfo.isZhubo = false;
					GlobalInfo.roomType = "ziyou";
					GlobalInfo.roomIp = room.ip;
					GlobalInfo.roomPort = room.port;
					GlobalInfo.subroomIp = room.mSubRoom.get(0).ip;
					GlobalInfo.subroomPort = room.mSubRoom.get(0).port;
					
					/* 新建一个Intent对象 */
		            Intent intent = new Intent();
		            //intent.putExtra("name","LeiPei");
		            /* 指定intent要启动的类 */
		            intent.setClass(ChatRoomListActivity.this, ChatRoomActivity.class);
		            /* 启动一个新的Activity */
		            ChatRoomListActivity.this.startActivity(intent);
		            /* 关闭当前的Activity */
		            ChatRoomListActivity.this.finish();
				}catch(Exception e){
					
				}
			}
		});
		thread.start();
	}

	public void onCreatePrivateZhubo(View view){
		
	}
	
	public void onCreatePrivateZiyou(View view){
		
	}
	
//	public void onJoin(View view){    
//		int pos = lv_chatroom.getSelectedItemPosition();
//		GLog.d("select pos:" + pos);
//    }  
}
