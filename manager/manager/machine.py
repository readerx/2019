# -*- coding: UTF-8 -*-

from flask import (
    Blueprint, flash, g, redirect, render_template, request, url_for
)
from werkzeug.exceptions import abort

from manager.auth import login_required
from manager.db import get_db


bp = Blueprint('machine', __name__)

@bp.route('/views', methods=('GET',))
@login_required
def views():
    db = get_db()
    machines = db.execute(
        'SELECT m.id, ip, cpu, mem, username, host'
        ' FROM machine m JOIN user u ON m.owner_id = u.id'
        ' ORDER BY ip DESC'
    ).fetchall()
    return render_template('machine/views.html', machines=machines)

@bp.route('/create', methods=('GET', 'POST'))
@login_required
def create():
    if request.method == 'POST':
        input_ip = request.form['inputIP']
        input_host = request.form['inputHost']
        input_owner = request.form['inputOwner']
        select_cpu = request.form['selectCPU']
        select_mem = request.form['selectMEM']
        error = None

        db = get_db()
        user = db.execute(
            'SELECT * FROM user WHERE username = ?', (input_owner,)
        ).fetchone()

        if user is None:
            error = 'Incorrect owner name.'

        if error is not None:
            flash(error)
        else:
            db.execute(
                'INSERT INTO machine (ip, host, cpu, mem, owner_id)'
                ' VALUES (?, ?, ?, ?, ?)',
                (input_ip, input_host, select_cpu, select_mem, user['id'])
            )
            db.commit()
            return redirect(url_for('machine.views'))

    return render_template('machine/create.html')

@bp.route('/<int:id>/delete', methods=('POST', 'GET'))
@login_required
def delete(id):
    db = get_db()
    db.execute('DELETE FROM machine WHERE id = ?', (id,))
    db.commit()
    return redirect(url_for('machine.views'))
