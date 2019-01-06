import sqlite3

conn = sqlite3.connect(r'E:\workspace\gitwork\2019\manager\instance\manager.sqlite')
cursor = conn.execute("SELECT username, password from user")
for row in cursor:
   print("ID = ", row[0])
   print("NAME = ", row[1])
   print("\n")

username = "admin"

user = conn.execute(
            'SELECT * FROM user WHERE username = ?', (username,)
        ).fetchone()

print(user)

conn.close()
