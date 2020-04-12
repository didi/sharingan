function Mutex() {
	this._isLocked = false;
	this._waiting = [];
}

export { Mutex }

Mutex.prototype.lock = function (cb) {
	if (this._isLocked) {
		this._waiting.push(cb);
	} else {
		this._isLocked = true;
		cb.call(this);
	}
};

Mutex.prototype.timedLock = function (ttl, cb) {
	if (!this._isLocked) {
		this._isLocked = true;
		return cb.call(this);
	}

	var timer, that = this;

	this._waiting.push(function () {
		clearTimeout(timer);

		if (!cb) {
			that.unlock();
			return;
		}

		cb.call(this);
		cb = null;
	});

	timer = setTimeout(function () {
		if (cb) {
			cb.call(this, new Error('Lock timed out'));
			cb = null;
		}
	}, ttl);
};

Object.defineProperty(Mutex.prototype, 'isLocked', {
	get: function () {
		return this._isLocked;
	}
});

Mutex.prototype.tryLock = function () {
	if (this._isLocked) {
		return false;
	}

	this._isLocked = true;
	return true;
};

Mutex.prototype.unlock = function () {
	if (!this._isLocked) {
		throw new Error('Mutex is not locked');
	}

	var waiter = this._waiting.shift();

	if (waiter) {
		waiter.call(this);
	} else {
		this._isLocked = false;
	}
};

Mutex.prototype.resetQueue = function() {
	this._waiting = [];
};
