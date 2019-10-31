import os
import re
import requests


def ppt(url):
    doc_id = re.findall('view/(.*).html', url)[0]
    url = "https://wenku.baidu.com/browse/getbcsurl?doc_id=" + doc_id + "&pn=1&rn=99999&type=ppt"
    html = requests.get(url).text

    lists = re.findall('{"zoom":"(.*?)","page"', html)
    for i in range(0, len(lists)):
        lists[i] = lists[i].replace("\\", '')

    os.mkdir(doc_id)
    for i in range(0, len(lists)):
        img = requests.get(lists[i]).content
        path = os.path.join(doc_id, "${index}.jpg".format(index=i))
        with open(path, 'wb') as m:
            m.write(img)
    print("dir: " + doc_id)


if __name__ == "__main__":
    ppt("view/45cb43a303020740be1e650e52ea551811a6c9da.html")
