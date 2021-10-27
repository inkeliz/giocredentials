package com.inkeliz.giocredentials;

import android.os.Bundle;
import android.accounts.AccountManager;
import android.content.Context;
import android.accounts.Account;
import android.util.Log;

public class cred_android  {

    public void add(Context context, String name, String password) {
        final AccountManager accountManager = AccountManager.get(context);
        final Account account = new Account(name, context.getPackageName());
        accountManager.addAccountExplicitly(account, password, null);
    }

    public String[][] list(Context context) {
        final AccountManager accountManager = AccountManager.get(context);
        final Account[] accounts = accountManager.getAccountsByType(context.getPackageName());

         final String[][] result = new String[accounts.length][];
         for (int i = 0; i < accounts.length; i++) {
             final String pass = accountManager.getPassword(accounts[i]);
             if (pass == null) {
               result[i] = new String[]{accounts[i].name, ""};
             } else {
               result[i] = new String[]{accounts[i].name, pass};
             }
         }

        return result;
    }

    public void remove(Context context, String name) {
        final AccountManager accountManager = AccountManager.get(context);
        final Account[] accounts = accountManager.getAccountsByType(context.getPackageName());

         for (int i = 0; i < accounts.length; i++) {
             if (accounts[i].name == name) {
                accountManager.removeAccountExplicitly(accounts[i]);
             }
         }
    }
}