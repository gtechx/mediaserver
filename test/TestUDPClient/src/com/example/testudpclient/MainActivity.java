package com.example.testudpclient;

import java.io.IOException;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.net.SocketException;
import java.net.UnknownHostException;
import java.util.HashMap;
import java.util.Map;

import android.app.Activity;
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
import android.widget.EditText;

public class MainActivity extends Activity {
	EditText txt_uid;
	EditText txt_ip;
	Button btn_joininput;
	Button btn_quitinput;
	Button btn_joinlistener;
	Button btn_quitlistener;
	
	boolean isquitlistener = true;
	boolean isquitinput = true;
	
	HashMap<String, AudioTrack> audioMap = new HashMap<String, AudioTrack>();
	
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
		setContentView(R.layout.activity_main);
		
		txt_ip = (EditText) findViewById(R.id.txt_ip);
		txt_uid = (EditText) findViewById(R.id.txt_uid);
		btn_joininput = (Button) findViewById(R.id.btn_joininput);
		btn_quitinput = (Button) findViewById(R.id.btn_quitinput);
		btn_joinlistener = (Button) findViewById(R.id.btn_joinlistener);
		btn_quitlistener = (Button) findViewById(R.id.btn_quitlistener);
		
		btn_joininput.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View view) {
				//
				if(!isquitlistener)
					return;
				isquitinput = false;
				startRecording();
				startUDPReading();
			}
		});
		
		btn_quitinput.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View view) {
				isquitinput = true;
			}
		});
		
		btn_joinlistener.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View view) {
				//
				if(!isquitinput)
					return;
				isquitlistener = false;
				startListenerUDPReading();
			}
		});
		
		btn_quitlistener.setOnClickListener(new OnClickListener() {
			@Override
			public void onClick(View view) {
				isquitlistener = true;
			}
		});
	}
	
	protected void startRecording(){
		Thread Rthread = new Thread(new Runnable() {
	        public void run() {
	        	InetAddress local = null;
	        	DatagramSocket dSocket = null;
	        	AudioRecord audioRecorder = null;
	        	long uid = Integer.parseInt(txt_uid.getText().toString());
	        	String ip = txt_ip.getText().toString();

				try {
					local = InetAddress.getByName(ip); 
				} catch (UnknownHostException e) {
					e.printStackTrace();
				}
				try {
					dSocket = new DatagramSocket(); 
				} catch (SocketException e) {
					e.printStackTrace();
				}
				byte[] data = new byte[13];
				byte[] uidbytes = getBytes(uid);
				System.arraycopy(uidbytes, 0, data, 4, 8);
				data[12] = 0;
				DatagramPacket dPacket = new DatagramPacket(data, 0, 13, local, 20001);
				try {
					dSocket.send(dPacket);
				} catch (IOException e) {
					e.printStackTrace();
				}
				
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
				while(isquitinput == false){
					int ret = audioRecorder.read(buffer, 0, buffer.length);   
                	
                	if(ret != AudioRecord.ERROR_INVALID_OPERATION && ret != AudioRecord.ERROR_BAD_VALUE && ret > 0){
                		int size = ret * 2;
                		byte[] sizebytes = getBytes(size);
                		uidbytes = getBytes(uid);
                		
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
				
				try{
					audioRecorder.stop();
					audioRecorder.release();
				}catch(Exception e){
					e.printStackTrace();
				}
				
				dSocket.close();
	        }
		});
		
		Rthread.start();
	}
	
	protected void startUDPReading(){
		Thread Rthread = new Thread(new Runnable() {
	        public void run() {
	        	InetAddress local = null;
	        	DatagramSocket dSocket = null;
	        	long uid = Integer.parseInt(txt_uid.getText().toString());
	        	String ip = txt_ip.getText().toString();

				try {
					local = InetAddress.getByName(ip); 
				} catch (UnknownHostException e) {
					e.printStackTrace();
				}
				try {
					dSocket = new DatagramSocket(); 
				} catch (SocketException e) {
					e.printStackTrace();
				}
				byte[] data = new byte[13];
				byte[] uidbytes = getBytes(uid);
				System.arraycopy(uidbytes, 0, data, 4, 8);
				data[12] = 0;
				DatagramPacket dPacket = new DatagramPacket(data, 0, 13, local, 30001);
				try {
					dSocket.send(dPacket);
				} catch (IOException e) {
					e.printStackTrace();
				}
				
				byte[] buffer = new byte[2048];
				DatagramPacket recvpacket = new DatagramPacket(buffer, 2048);

				while(isquitinput != true){
					try{
					dSocket.receive(recvpacket);
					}catch(Exception e){
						e.printStackTrace();
					}
					
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
				}
				
				dSocket.close();
				for (Map.Entry<String, AudioTrack> entry : audioMap.entrySet()) {
					entry.getValue().stop();
					entry.getValue().release();
			    }
				audioMap.clear();
	        }
		});
		
		Rthread.start();
	}
	
	protected void startListenerUDPReading(){
		Thread Rthread = new Thread(new Runnable() {
	        public void run() {
	        	InetAddress local = null;
	        	DatagramSocket dSocket = null;
	        	long uid = Integer.parseInt(txt_uid.getText().toString());
	        	String ip = txt_ip.getText().toString();

				try {
					local = InetAddress.getByName(ip); 
				} catch (UnknownHostException e) {
					e.printStackTrace();
				}
				try {
					dSocket = new DatagramSocket(); 
				} catch (SocketException e) {
					e.printStackTrace();
				}
				byte[] data = new byte[13];
				byte[] uidbytes = getBytes(uid);
				System.arraycopy(uidbytes, 0, data, 4, 8);
				data[12] = 0;
				DatagramPacket dPacket = new DatagramPacket(data, 0, 13, local, 30001);
				try {
					dSocket.send(dPacket);
				} catch (IOException e) {
					e.printStackTrace();
				}
				
				byte[] buffer = new byte[2048];
				DatagramPacket recvpacket = new DatagramPacket(buffer, 2048);

				while(isquitlistener != true){
					try{
					dSocket.receive(recvpacket);
					}catch(Exception e){
						e.printStackTrace();
					}
					
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
				}
				
				dSocket.close();

				for (Map.Entry<String, AudioTrack> entry : audioMap.entrySet()) {
					entry.getValue().stop();
					entry.getValue().release();
			    }
				audioMap.clear();
	        }
		});
		
		Rthread.start();
	}
}
