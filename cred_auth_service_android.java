package com.inkeliz.giocredentials;

import android.app.Service;
import android.content.Intent;
import android.os.IBinder;

public class cred_auth_service_android extends Service {
	private static cred_auth_android auth;

	@Override
	public IBinder onBind(Intent intent) {
		IBinder binder = null;
		if (intent.getAction().equals(android.accounts.AccountManager.ACTION_AUTHENTICATOR_INTENT)) {
			binder = getAuthenticator().getIBinder();
		}
		return binder;
	}

	private cred_auth_android getAuthenticator() {
		if (null == cred_auth_service_android.auth) {
			cred_auth_service_android.auth = new cred_auth_android(this);
		}
		return cred_auth_service_android.auth;
	}
}