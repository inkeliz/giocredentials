GioCredentials
--

----

**What is GioCredentials?** 

It's a small library to store (and retrieve) usernames and passwords (or equivalent type of data).

**Where the credentials are stored?**

| OS | Default | Fallback/Unsafe |
| --- | ----------- | --- |
| **Web Browser** | |
| WebAssembly | [Credential Management API](https://developer.mozilla.org/en-US/docs/Web/API/Credential_Management_API) | [Web Storage API](https://developer.mozilla.org/en-US/docs/Web/API/Web_Storage_API) |
|||
| **Desktop** | |
| Windows 10+ | [Wincred API](https://docs.microsoft.com/en-us/windows/win32/api/wincred/) | None ❌ |
| Linux KDE | Unsupported ❌| Unsupported ❌ |
| Linux Gnome | Unsupported ❌| Unsupported ❌ |
| MacOS |  Unsupported ❌ | Unsupported ❌ |
|||
| **Mobile** | |
| Android 5.0+ | [AccountManager API](https://developer.android.com/reference/android/accounts/AccountManager) | None ❌ |
| iOS | Unsupported ❌ | Unsupported ❌ |

**How secure it is?**

It varies due to the technology used on each OS. For Windows (and Web-Browsers on Windows): other apps on the same machine can retrieve the credentials (it's not per-application). On Android: it's insecure if the device is rooted. It's important to notice that other libraries (including dynamic/external libraries) can retrieve the credentials (such as other JS running on the same webpage, plugins or other library that is bundled with your app).  Some API might not be available, that is why we provide the `Unsafe` function. For instance Safari 14 doesn't have full support for Credential Management API, Windows 7 (and earlier) doesn't have Wincred API. The `Unsafe` uses what is available to store, even if it's not exclusively design to store password. It's important to note that web-browsers vendors are know that WebStorage API could be used to store ["sensitivity data"](https://html.spec.whatwg.org/multipage/webstorage.html#sensitivity-of-data).

Vulnerabilities and catches and gotchas:
---

- ***Timing-Attack on table-lookup***:   
**JS and Android** (Default and Unsafe): the `Base64.StdEncode` is used to store the password, since `Password` field is intended to be string.

  
- ***Sub-Domain attack***:   
**JS** (Default and Unsafe): any subdomain, under the same domain, could access the credentials.

  
- ***`Remove()` function isn't reliable***:    
**JS** (Default): is impossible to delete credentials. It's a limitation of Credentials Management API, it acts like a no-op function.

  
- ***`Credentials.Username` and `Credentials.Password` can't be longer than 512 bytes:***    
**Windows**: doesn't support long password/username. In the future, GioCredentials will split credentials into multiple parts.
  

- ***Credentials can be read by any app***:   
**Windows**: any other executable on the machine (installed on the same windows-user) can read all credentials.
  

- ***OS-Specific features isn't available***:   
Some APIs are more powerful than others, some APIs could handle cryptography keys natively or can store additional data. However, none of those features are expose, since GioCredentials is aimed to work on any platform.
  

Android (Gio)
--------

You must define some `permissions` and `service` and a new file `xml/authenticator.xml`:

> ***res/xml/authenticator.xml***:
```
<account-authenticator xmlns:android="http://schemas.android.com/apk/res/android"
                       android:accountType="**Your App ID (com.your.app)**"
                       android:icon="@mipmap/ic_launcher"
                       android:smallIcon="@mipmap/ic_launcher"/>
```

Replace `**Your App ID (com.your.app)**` with your custom ID.

> ***manifest.xml***:
```
<uses-permission android:name="android.permission.GET_ACCOUNTS" />
<uses-permission android:name="android.permission.MANAGE_ACCOUNTS" />
<uses-permission android:name="android.permission.AUTHENTICATE_ACCOUNTS" />
<uses-permission android:name="android.permission.USE_CREDENTIALS" />
```
```
<application>
    <!-- ... ->
    
    <service android:name="com.inkeliz.giocredentials.cred_auth_service_android">
	    <intent-filter>
		    <action android:name="android.accounts.AccountAuthenticator"/>
		</intent-filter>
		<meta-data
			android:name="android.accounts.AccountAuthenticator"
			android:resource="@xml/authenticator" />
	</service>
</application>
```

***There's no way to set custom `manifest` or files, when using `gogio`, so you must patch manually***:

```
diff --git "a/cmd/gogio/androidbuild.go" "b/cmd/gogio/androidbuild.go"
index fc21d40..2de8dd0 100644
--- "a/cmd/gogio/androidbuild.go"
+++ "b/cmd/gogio/androidbuild.go"
@@ -348,7 +348,9 @@ func exeAndroid(tmpDir string, tools *androidTools, bi *buildInfo, extraJars, pe
 	resDir := filepath.Join(tmpDir, "res")
 	valDir := filepath.Join(resDir, "values")
 	v21Dir := filepath.Join(resDir, "values-v21")
-	for _, dir := range []string{valDir, v21Dir} {
+	xmlDir := filepath.Join(resDir, "xml")
+	xmlDirV32 := filepath.Join(resDir, "xml-v32")
+	for _, dir := range []string{valDir, v21Dir, xmlDir, xmlDirV32} {
 		if err := os.MkdirAll(dir, 0755); err != nil {
 			return err
 		}
@@ -357,6 +359,7 @@ func exeAndroid(tmpDir string, tools *androidTools, bi *buildInfo, extraJars, pe
 	if _, err := os.Stat(bi.iconPath); err == nil {
 		err := buildIcons(resDir, bi.iconPath, []iconVariant{
 			{path: filepath.Join("mipmap-hdpi", "ic_launcher.png"), size: 72},
+			{path: filepath.Join("drawable", "ic_launcher.png"), size: 72},
 			{path: filepath.Join("mipmap-xhdpi", "ic_launcher.png"), size: 96},
 			{path: filepath.Join("mipmap-xxhdpi", "ic_launcher.png"), size: 144},
 			{path: filepath.Join("mipmap-xxxhdpi", "ic_launcher.png"), size: 192},
@@ -374,6 +377,11 @@ func exeAndroid(tmpDir string, tools *androidTools, bi *buildInfo, extraJars, pe
 	if err != nil {
 		return err
 	}
+	ioutil.WriteFile(filepath.Join(xmlDir, "authenticator.xml"), []byte(`<?xml version="1.0" encoding="utf-8"?>
+<account-authenticator xmlns:android="http://schemas.android.com/apk/res/android"
+                       android:accountType="`+bi.appID+`"
+                       android:icon="@drawable/ic_launcher"
+                       android:smallIcon="@drawable/ic_launcher"/>`), 0660)
 	resZip := filepath.Join(tmpDir, "resources.zip")
 	aapt2 := filepath.Join(tools.buildtools, "aapt2")
 	_, err = runCmd(exec.Command(
@@ -415,6 +423,10 @@ func exeAndroid(tmpDir string, tools *androidTools, bi *buildInfo, extraJars, pe
 	android:versionCode="{{.Version}}"
 	android:versionName="1.0.{{.Version}}">
 	<uses-sdk android:minSdkVersion="{{.MinSDK}}" android:targetSdkVersion="{{.TargetSDK}}" />
+<uses-permission android:name="android.permission.GET_ACCOUNTS" />
+<uses-permission android:name="android.permission.MANAGE_ACCOUNTS" />
+<uses-permission android:name="android.permission.AUTHENTICATE_ACCOUNTS" />
+<uses-permission android:name="android.permission.USE_CREDENTIALS" />
 {{range .Permissions}}	<uses-permission android:name="{{.}}"/>
 {{end}}{{range .Features}}	<uses-feature android:{{.}} android:required="false"/>
 {{end}}	<application {{.IconSnip}} android:label="{{.AppName}}">
@@ -428,6 +440,14 @@ func exeAndroid(tmpDir string, tools *androidTools, bi *buildInfo, extraJars, pe
 				<category android:name="android.intent.category.LAUNCHER" />
 			</intent-filter>
 		</activity>
+		<service android:name="com.inkeliz.giocredentials.cred_auth_service_android">
+			<intent-filter>
+				<action android:name="android.accounts.AccountAuthenticator"/>
+			</intent-filter>
+			<meta-data
+				android:name="android.accounts.AccountAuthenticator"
+				android:resource="@xml/authenticator" />
+		</service>
 	</application>
 </manifest>`)
 	var manifestBuffer bytes.Buffer
```

Make sure to build with `-appId com.your.app`, it must be unique and must be the same defined in `android:accountType`. In any case, without Gio, you must declare: