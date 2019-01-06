# -*- coding: UTF-8 -*-

from flask import (
    Blueprint, flash, g, redirect, render_template, request, url_for, session
)
from werkzeug.security import check_password_hash, generate_password_hash

from manager.db import get_db
from common.util import check_pasword


bp = Blueprint('index', __name__)

@bp.route('/', methods=('GET', 'POST'))
def index():
    if request.method == 'POST':
        username = request.form['username']
        password = request.form['password']

        print(username)

        db = get_db()
        error = None
        user = db.execute(
            'SELECT * FROM user WHERE username = ?', (username,)
        ).fetchone()

        if user is None:
            error = 'Incorrect username.'
        elif not check_pasword(user['password'], password):
            error = 'Incorrect password.'

        print("error: ", error)

        if error is None:
            session.clear()
            session['user_id'] = user['id']
            return redirect(url_for('machine.views'))

        flash(error)

    return render_template('index.html')
