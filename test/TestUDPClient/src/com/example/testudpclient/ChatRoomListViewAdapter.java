package com.example.testudpclient;

import java.util.ArrayList;

import com.example.testudpclient.ChatRoomListActivity.RoomInfo;
import com.giant.sdk.log.GLog;

import android.app.Activity;
import android.content.Context;
import android.content.Intent;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.view.View.OnClickListener;
import android.widget.BaseAdapter;
import android.widget.Button;
import android.widget.TextView;

public class ChatRoomListViewAdapter extends BaseAdapter {
	
	ArrayList<RoomInfo> mRoomList;
	Context context;
	LayoutInflater mInflater;
	
	public class ViewHolder {
		public TextView tv_chatroom;
		public TextView tv_type;
	}
	
	public ChatRoomListViewAdapter(Context con, ArrayList<RoomInfo> roomlist){
		mRoomList = roomlist;
		context = con;
		mInflater = LayoutInflater.from(context);
	}
	
	@Override
	public int getCount() {
		// TODO Auto-generated method stub
		return mRoomList.size();
	}

	@Override
	public Object getItem(int position) {
		// TODO Auto-generated method stub
		return mRoomList.get(position);
	}

	@Override
	public long getItemId(int position) {
		// TODO Auto-generated method stub
		return position;
	}

	@Override
	public View getView(int position, View convertView, ViewGroup parent) {
		// TODO Auto-generated method stub
		ViewHolder holder = null;
		final RoomInfo rinfo = mRoomList.get(position);
		if (convertView == null) {
			convertView = mInflater.inflate(R.layout.item_chatroom, null);
			
			holder = new ViewHolder();
			holder.tv_chatroom = (TextView) convertView.findViewById(R.id.tv_chatroom);
			holder.tv_type = (TextView) convertView.findViewById(R.id.tv_type);
			Button btn_join = (Button)convertView.findViewById(R.id.btn_join);
			final int pos = position;
			btn_join.setOnClickListener(new OnClickListener(){
				@Override
				public void onClick(View view) {
					GLog.d("info:" + rinfo.type + " room ip=" + rinfo.room.ip + " port=" + rinfo.room.port);
					GLog.d("info:" + rinfo.type + " subroom ip=" + rinfo.room.mSubRoom.get(0).ip + " port=" + rinfo.room.mSubRoom.get(0).port);
					
//					if(rinfo.type == "zhubo"){
//						GlobalInfo.isZhubo = false;
//						GlobalInfo.roomType = rinfo.type;
//						GlobalInfo.roomIp = rinfo.room.ip;
//						GlobalInfo.roomPort = rinfo.room.port;
//						GlobalInfo.subroomIp = rinfo.room.mSubRoom.get(0).ip;
//						GlobalInfo.subroomPort = rinfo.room.mSubRoom.get(0).port;
//					}else{
//						
//					}
					
					GlobalInfo.isZhubo = false;
					GlobalInfo.roomType = rinfo.type;
					GlobalInfo.roomIp = rinfo.room.ip;
					GlobalInfo.roomPort = rinfo.room.port;
					GlobalInfo.subroomIp = rinfo.room.mSubRoom.get(0).ip;
					GlobalInfo.subroomPort = rinfo.room.mSubRoom.get(0).port;
					
					/* 新建一个Intent对象 */
	                Intent intent = new Intent();
	                //intent.putExtra("name","LeiPei");
	                /* 指定intent要启动的类 */
	                intent.setClass(context, ChatRoomActivity.class);
	                /* 启动一个新的Activity */
	                context.startActivity(intent);
	                /* 关闭当前的Activity */
	                ((Activity)context).finish();
				}
			});
			convertView.setTag(holder);
		} else {
			holder = (ViewHolder) convertView.getTag();
		}
		
		holder.tv_chatroom.setText(rinfo.room.id);
		holder.tv_type.setText(rinfo.type);
		
		//holder.textView.setText(mContentList.get(position));
		return convertView;
	}

}
