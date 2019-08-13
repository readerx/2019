import urllib.request
import urllib.parse
import json
import re


def read_view(doc_id):
    url = 'https://wenku.baidu.com/view/{id}.html'.format(id=doc_id)
    response = urllib.request.urlopen(url)
    status = response.getcode()
    if status != 200:
        raise Exception("request view status error, status {status}".format(status=status))

    content = response.read()
    return content.decode('ISO-8859-1')


def parse_view(data):
    pattern = re.compile(r'WkInfo.htmlUrls = \'(.*)\';')
    result = pattern.findall(data)
    if len(result) != 1:
        raise Exception("pares view error, result size {size}".format(size=len(result)))
    return result[0].replace('\\x22', '"')


def read_page_json(data):
    for info in data:
        url = info['pageLoadUrl'].replace('\\', '')
        data = read_data(url)
        format_data(info['pageIndex'], data)


def read_data(url):
    response = urllib.request.urlopen(url)
    status = response.getcode()
    if status > 300 or status < 200:
        raise Exception("request data status error, status {status}".format(status=status))
    content = response.read()
    return content.decode('utf-8')


def format_data(page, data):
    pattern = re.compile(r'wenku_\d+\((.*)\)')
    result = pattern.findall(data)
    if len(result) != 1:
        raise Exception("format data error, result size {size}".format(size=len(result)))

    j = json.loads(result[0])
    for b in j['body']:
        print(b['c'], end="")

        if b['ps'] is not None:
            print("\n")


if __name__ == "__main__":
    view = read_view("136da9c702d276a200292ea0")
    pages = parse_view(view)
    js = json.loads(pages)
    read_page_json(js["json"])
