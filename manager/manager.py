#!/usr/bin/python
# -*- coding: UTF-8 -*-

from flask import Flask
from flask import render_template


app = Flask(__name__)

@app.route('/machines')
def machines():
    return render_template('hello.html', name="testxx")
