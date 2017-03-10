package com.example.testudpclient;

import java.io.IOException;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.net.SocketException;
import java.net.UnknownHostException;
import java.util.HashMap;
import java.util.Map;

import com.giant.sdk.log.GLog;

import android.app.Activity;
import android.content.Intent;
import android.media.AudioFormat;
import android.media.AudioManager;
import android.media.AudioRecord;
import android.media.AudioTrack;
import android.media.MediaRecorder.AudioSource;
import android.os.Bundle;
import android.util.Log;
import android.view.View;
import android.view.View.OnClickListener;
import android.widget.Button;

public class ChatRoomActivity extends Activity {
	Button btn_quit;
	
	HashMap<String, AudioTrack> audioMap = new HashMap<String, AudioTrack>();
	boolean isquit = false;
	
	Object udpSenderLock = new Object();
	Object udpReaderLock = new Object();
	Thread udpReadingThread;
	Thread udpSendingThread;
	DatagramSocket udpReadingSocket;
	DatagramSocket udpSendingSocket;
	
	public static long getLong(byte[] bytes, int ofs) {
		return (0xffL & (long) bytes[0 + ofs]) | (0xff00L & ((long) bytes[1 + ofs] << 8)) | (0xff0000L & ((long) bytes[2 + ofs] << 16)) | (0xff000000L & ((long) bytes[3 + ofs] << 24))
				| (0xff00000000L & ((long) bytes[4 + ofs] << 32)) | (0xff0000000000L & ((long) bytes[5 + ofs] << 40)) | (0xff000000000000L & ((long) bytes[6 + ofs] << 48))
				| (0xff00000000000000L & ((long) bytes[7 + ofs] << 56));
	}
	
	public static int getInt(byte[] bytes) {
		return (0xff & bytes[0]) | (0xff00 & (bytes[1] << 8)) | (0xff0000 & (bytes[2] << 16)) | (0xff000000 & (bytes[3] << 24));
	}
	
	public static byte[] getBytes(int data) {
		byte[] bytes = new byte[4];
		bytes[0] = (byte) (data & 0xff);
		bytes[1] = (byte) ((data & 0xff00) >> 8);
		bytes[2] = (byte) ((data & 0xff0000) >> 16);
		bytes[3] = (byte) ((data & 0xff000000) >> 24);
		return bytes;
	}

	/**
	 * long转化为byte数组
	 * 
	 * @param data
	 * @return
	 */
	public static byte[] getBytes(long data) {
		byte[] bytes = new byte[8];
		bytes[0] = (byte) (data & 0xff);
		bytes[1] = (byte) ((data >> 8) & 0xff);
		bytes[2] = (byte) ((data >> 16) & 0xff);
		bytes[3] = (byte) ((data >> 24) & 0xff);
		bytes[4] = (byte) ((data >> 32) & 0xff);
		bytes[5] = (byte) ((data >> 40) & 0xff);
		bytes[6] = (byte) ((data >> 48) & 0xff);
		bytes[7] = (byte) ((data >> 56) & 0xff);
		return bytes;
	}
	
	@Override
	protected void onCreate(Bundle savedInstanceState) {
		super.onCreate(savedInstanceState);
		setContentView(R.layout.activity_chatroom);
		
		btn_quit = (Button)this.findViewById(R.id.btn_quit);
		
		btn_quit.setOnClickListener(new OnClickListener(){
			@Override
			public void onClick(View view) {
				try{
				isquit = true;
				
				if(udpSendingSocket != null)
					udpSendingSocket.close();
				if(udpReadingSocket != null)
					udpReadingSocket.close();
				
				if(GlobalInfo.roomType == "zhubo"){
					synchronized(udpSenderLock) {
						
					}
					synchronized(udpReaderLock) {
						
					}
				}else{
					synchronized(udpReaderLock) {
						
					}
				}
				
				/* 新建一个Intent对象 */
                Intent intent = new Intent();
                //intent.putExtra("name","LeiPei");
                /* 指定intent要启动的类 */
                intent.setClass(ChatRoomActivity.this, ChatRoomListActivity.class);
                /* 启动一个新的Activity */
                ChatRoomActivity.this.startActivity(intent);
                /* 关闭当前的Activity */
                ChatRoomActivity.this.finish();
				}catch(Exception e){
					e.printStackTrace();
				}
			}
		});
		
		Thread thread = new Thread(new Runnable(){
			public void run(){
				if(GlobalInfo.roomType == "zhubo"){
					joinBroadcastRoom();
				}else{
					joinReceiveRoom();
					joinBroadcastRoom();
				}
			}
		});
		thread.start();
	}
	
	protected void joinReceiveRoom(){
		InetAddress local = null;
		udpSendingSocket = null;
    	GLog.d("login to receive server...");
    	try {
			local = InetAddress.getByName(GlobalInfo.roomIp); 
		} catch (UnknownHostException e) {
			e.printStackTrace();
		}
		try {
			udpSendingSocket = new DatagramSocket(); 
		} catch (SocketException e) {
			e.printStackTrace();
		}
		byte[] data = new byte[13];
		byte[] uidbytes = getBytes(GlobalInfo.sessionId);
		System.arraycopy(uidbytes, 0, data, 4, 8);
		data[12] = 0;
		DatagramPacket dPacket = new DatagramPacket(data, 0, 13, local, GlobalInfo.roomPort);
		try {
			udpSendingSocket.send(dPacket);
		} catch (IOException e) {
			e.printStackTrace();
		}
		
		byte[] buffer = new byte[2048];
		DatagramPacket recvpacket = new DatagramPacket(buffer, 2048);

		try{
			udpSendingSocket.receive(recvpacket);
		}catch(Exception e){
			e.printStackTrace();
		}
		
		//int datasize = getInt(buffer);
		//long sessionid = getLong(buffer, 4);
		short type = (short)(buffer[12] & 0xff);

		if(type == 200){
			//join rs room success
			GLog.d("login to receive server success");
			startRecording(local, udpSendingSocket);
		}
		else{
			GLog.d("login to receive server failed:" + type);
		}
		
		//dSocket.close();
	}
	
	protected void joinBroadcastRoom(){
		InetAddress local = null;
		udpReadingSocket = null;
    	GLog.d("login to broadcast server...");
    	try {
			local = InetAddress.getByName(GlobalInfo.subroomIp); 
		} catch (UnknownHostException e) {
			e.printStackTrace();
		}
		try {
			udpReadingSocket = new DatagramSocket(); 
		} catch (SocketException e) {
			e.printStackTrace();
		}
		byte[] data = new byte[13];
		byte[] uidbytes = getBytes(GlobalInfo.sessionId);
		System.arraycopy(uidbytes, 0, data, 4, 8);
		data[12] = 0;
		DatagramPacket dPacket = new DatagramPacket(data, 0, 13, local, GlobalInfo.subroomPort);
		try {
			udpReadingSocket.send(dPacket);
		} catch (IOException e) {
			e.printStackTrace();
		}
		
		byte[] buffer = new byte[2048];
		DatagramPacket recvpacket = new DatagramPacket(buffer, 2048);

		try{
			udpReadingSocket.receive(recvpacket);
		}catch(Exception e){
			e.printStackTrace();
		}
		
		//int datasize = getInt(buffer);
		//long sessionid = getLong(buffer, 4);
		short type = (short)(buffer[12] & 0xff);

		if(type == 200){
			//join rs room success
			GLog.d("login to broadcast server success");
			startUDPReading(local, udpReadingSocket);
		}
		else{
			GLog.d("login to broadcast server failed:" + type);
		}
		
		//dSocket.close();
	}
	
	protected void startRecording(final InetAddress local, final DatagramSocket dSocket){
		udpSendingThread = new Thread(new Runnable() {
	        public void run() {
	        	//InetAddress local = null;
	        	//DatagramSocket dSocket = null;
	        	
	        	AudioRecord audioRecorder = null;
	        	//long uid = Integer.parseInt(txt_uid.getText().toString());
	        	//String ip = txt_ip.getText().toString();

//				try {
//					local = InetAddress.getByName(GlobalInfo.roomIp); 
//				} catch (UnknownHostException e) {
//					e.printStackTrace();
//				}
//				try {
//					dSocket = new DatagramSocket(); 
//				} catch (SocketException e) {
//					e.printStackTrace();
//				}
//				byte[] data = new byte[13];
//				byte[] uidbytes = getBytes(GlobalInfo.sessionId);
//				System.arraycopy(uidbytes, 0, data, 4, 8);
//				data[12] = 0;
//				DatagramPacket dPacket = new DatagramPacket(data, 0, 13, local, GlobalInfo.roomPort);
//				try {
//					dSocket.send(dPacket);
//				} catch (IOException e) {
//					e.printStackTrace();
//				}
				
				try{
					audioRecorder = new AudioRecord(AudioSource.MIC, 8000, AudioFormat.CHANNEL_IN_MONO, AudioFormat.ENCODING_PCM_16BIT, 1280);
				} catch (Exception e){
					e.printStackTrace();
					return;
				}
				
				try{
					audioRecorder.startRecording();
				}catch(Exception e){
					e.printStackTrace();
					return;
				}
				
				short[] buffer = new short[320];
				byte[] sendbuffer = new byte[653];
				GLog.d("start sending data ro receive server...");
				synchronized(udpSenderLock) {
					while(isquit == false){
						int ret = audioRecorder.read(buffer, 0, buffer.length);   
	                	
	                	if(ret != AudioRecord.ERROR_INVALID_OPERATION && ret != AudioRecord.ERROR_BAD_VALUE && ret > 0){
	                		int size = ret * 2;
	                		byte[] sizebytes = getBytes(size);
	                		byte[] uidbytes = getBytes(GlobalInfo.sessionId);
	                		
	                		System.arraycopy(sizebytes, 0, sendbuffer, 0, 4);
	                		System.arraycopy(uidbytes, 0, sendbuffer, 4, 8);
	                		sendbuffer[12] = 2;
	                		
	                		byte[] byteBuff = new byte[size];
		        			for(int i = 0; i < ret; ++i){
		        				short dat = buffer[i];

		        				byteBuff[i * 2 + 0] = (byte) (dat & 0xff);
		        				byteBuff[i * 2 + 1] = (byte) ((dat & 0xff00) >> 8);
		        			}
	                		System.arraycopy(byteBuff, 0, sendbuffer, 13, size);
	                		DatagramPacket sendpak = new DatagramPacket(sendbuffer, 0, sendbuffer.length, local, 20001);
							try {
								dSocket.send(sendpak);
							} catch (IOException e) {
								e.printStackTrace();
							}
	                	}
	                	else{
	                		Log.d("TESTUDPCLIENT", "AudioRecord.ERROR_INVALID_OPERATION or AudioRecord.ERROR_BAD_VALUE Get is " + ret);
	                	}
					}
				}
				GLog.d("quit receive server...");
				try{
					audioRecorder.stop();
					audioRecorder.release();
				}catch(Exception e){
					e.printStackTrace();
				}
				
				dSocket.close();
	        }
		});
		
		udpSendingThread.start();
	}
	
	protected void startUDPReading(final InetAddress local, final DatagramSocket dSocket){
		udpReadingThread = new Thread(new Runnable() {
	        public void run() {
	        	//InetAddress local = null;
	        	//DatagramSocket dSocket = null;
	        	//long uid = Integer.parseInt(txt_uid.getText().toString());

//				try {
//					local = InetAddress.getByName(GlobalInfo.subroomIp); 
//				} catch (UnknownHostException e) {
//					e.printStackTrace();
//				}
//				try {
//					dSocket = new DatagramSocket(); 
//				} catch (SocketException e) {
//					e.printStackTrace();
//				}
//				byte[] data = new byte[13];
//				byte[] uidbytes = getBytes(GlobalInfo.sessionId);
//				System.arraycopy(uidbytes, 0, data, 4, 8);
//				data[12] = 0;
//				DatagramPacket dPacket = new DatagramPacket(data, 0, 13, local, GlobalInfo.subroomPort);
//				try {
//					dSocket.send(dPacket);
//				} catch (IOException e) {
//					e.printStackTrace();
//				}
				
				byte[] buffer = new byte[2048];
				DatagramPacket recvpacket = new DatagramPacket(buffer, 2048);
				GLog.d("start receiving data from broadcast server...");
				synchronized(udpReaderLock) {
					while(isquit == false){
						try{
							dSocket.receive(recvpacket);
							
							int datasize = getInt(buffer);
							long ruid = getLong(buffer, 4);
							//byte type = buffer[12];
							AudioTrack audio;
							String struid = "" + ruid;
							if(audioMap.containsKey(struid)){
								audio = audioMap.get(struid);
							}
							else{
								audio = new AudioTrack(
									     AudioManager.STREAM_MUSIC,
									     8000,
									     AudioFormat.CHANNEL_OUT_MONO,
									     AudioFormat.ENCODING_PCM_16BIT,
									     4096,
									     AudioTrack.MODE_STREAM
									     );
								audio.play();
								audioMap.put(struid, audio);
							}
							byte[] databuffer = new byte[datasize];
							System.arraycopy(buffer, 13, databuffer, 0, datasize);
							
							audio.write(databuffer, 0, datasize);
						}catch(Exception e){
							e.printStackTrace();
						}
					}
				}
				GLog.d("quit broadcast server...");
				dSocket.close();
				for (Map.Entry<String, AudioTrack> entry : audioMap.entrySet()) {
					entry.getValue().stop();
					entry.getValue().release();
			    }
				audioMap.clear();
	        }
		});
		
		udpReadingThread.start();
	}
}
