#!/usr/bin/python
# -*- coding: UTF-8 -*-

class Machine:
    """
    machine结构定义
    """

    def __init__(self, ip, user, cpu, mem, host):
        self.ip = ip
        self.user = user
        self.cpu = cpu
        self.mem = mem
        self.host = host
