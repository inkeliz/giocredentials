package giocredentials

import (
	"encoding/hex"
	"gioui.org/app"
	"git.wow.st/gmp/jni"
	"math"
	"sync"
)

//go:generate javac -source 8 -target 8 -bootclasspath $ANDROID_HOME\platforms\android-29\android.jar -d $TEMP\giocred\classes cred_android.java cred_auth_android.java cred_auth_service_android.java
//go:generate jar cf cred_android.jar -C $TEMP\giocred\classes .

type config struct {
	java struct {
		class jni.Class
		obj   jni.Object
		sync.Mutex
	}
}

func (m *Manager) view() (creds []*Credential, err error) {
	err = jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		cls, obj, err := m.getJavaCred()
		if err != nil {
			return err
		}

		mid := jni.GetMethodID(env, cls, "list", "(Landroid/content/Context;)[[Ljava/lang/String;")
		values := []jni.Value{
			jni.Value(jni.Object(app.AppContext())),
		}
		resp, err := jni.CallObjectMethod(env, obj, mid, values...)
		if err != nil {
			return err
		}

		creds = make([]*Credential, 0, 16)
		for i := 0; i < math.MaxInt16; i++ {
			array, err := jni.GetObjectArrayElement(env, jni.ObjectArray(resp), jni.Size(i))
			if err != nil {
				break
			}

			user, _ := jni.GetObjectArrayElement(env, jni.ObjectArray(array), 0)
			pass, _ := jni.GetObjectArrayElement(env, jni.ObjectArray(array), 1)

			cred := &Credential{}
			cred.Username = jni.GoString(env, jni.String(user))
			cred.Password, err = hex.DecodeString(jni.GoString(env, jni.String(pass)))

			creds = append(creds, cred)
		}

		return nil
	})

	return creds, err
}

func (m *Manager) add(cred *Credential) (err error) {
	err = jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		cls, obj, err := m.getJavaCred()
		if err != nil {
			return err
		}

		mid := jni.GetMethodID(env, cls, "add", "(Landroid/content/Context;Ljava/lang/String;Ljava/lang/String;)V")
		values := []jni.Value{
			jni.Value(jni.Object(app.AppContext())),
			jni.Value(jni.JavaString(env, cred.Username)),
			jni.Value(jni.JavaString(env, hex.EncodeToString(cred.Password))),
		}
		if err := jni.CallVoidMethod(env, obj, mid, values...); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (m *Manager) remove(cred *Credential) (err error) {
	err = jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		cls, obj, err := m.getJavaCred()
		if err != nil {
			return err
		}

		mid := jni.GetMethodID(env, cls, "remove", "(Landroid/content/Context;Ljava/lang/String;)V")
		values := []jni.Value{
			jni.Value(jni.Object(app.AppContext())),
			jni.Value(jni.JavaString(env, cred.Username)),
		}
		if err := jni.CallVoidMethod(env, obj, mid, values...); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (m *Manager) getJavaCred() (class jni.Class, obj jni.Object, err error) {
	m.java.Lock()
	defer m.java.Unlock()

	if m.java.obj != 0 && m.java.class != 0 {
		return m.java.class, m.java.obj, nil
	}

	err = jni.Do(jni.JVMFor(app.JavaVM()), func(env jni.Env) error {
		// Get the GioView object
		obj := jni.Object(app.AppContext())
		cls := jni.GetObjectClass(env, obj)

		mid := jni.GetMethodID(env, cls, "getClassLoader", "()Ljava/lang/ClassLoader;")
		obj, err = jni.CallObjectMethod(env, obj, mid)
		if err != nil {
			return err
		}

		// Run findClass() to get the custom class
		cls = jni.GetObjectClass(env, obj)
		mid = jni.GetMethodID(env, cls, "findClass", "(Ljava/lang/String;)Ljava/lang/Class;")
		clso, err := jni.CallObjectMethod(env, obj, mid, jni.Value(jni.JavaString(env, `com.inkeliz.giocredentials.cred_android`)))
		if err != nil {
			return err
		}

		// We need to create an GlobalRef of our class, otherwise we can't manipulate that afterwards.
		m.java.class = jni.Class(jni.NewGlobalRef(env, clso))

		// Create a new Object from our class.
		mid = jni.GetMethodID(env, m.java.class, "<init>", `()V`)
		obj, err = jni.NewObject(env, m.java.class, mid)
		if err != nil {
			return err
		}

		// We need to create an GlobalRef of our object.
		m.java.obj = jni.Object(jni.NewGlobalRef(env, obj))
		return nil
	})

	if err != nil {
		return 0, 0, err
	}

	return m.java.class, m.java.obj, nil
}
