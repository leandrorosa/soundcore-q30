#import <IOBluetooth/IOBluetooth.h>
#include "rfcomm_darwin.h"

@interface Q30Channel : NSObject <IOBluetoothRFCOMMChannelDelegate>
@property(strong) IOBluetoothRFCOMMChannel *channel;
@property(strong) NSMutableData *buf;
@property BOOL closed;
@property IOReturn openStatus;
@property BOOL opened;
@end

@implementation Q30Channel
- (void)rfcommChannelOpenComplete:(IOBluetoothRFCOMMChannel *)ch status:(IOReturn)status {
	self.openStatus = status;
	self.opened = YES;
}
- (void)rfcommChannelData:(IOBluetoothRFCOMMChannel *)ch data:(void *)p length:(size_t)n {
	[self.buf appendBytes:p length:n];
}
- (void)rfcommChannelClosed:(IOBluetoothRFCOMMChannel *)ch {
	self.closed = YES;
}
@end

static void spinUntil(NSDate *deadline, BOOL (^done)(void)) {
	while (!done() && [deadline timeIntervalSinceNow] > 0) {
		CFRunLoopRunInMode(kCFRunLoopDefaultMode, 0.05, true);
	}
}

// IOBluetooth's CoreBluetooth backend powers on asynchronously and silently
// drops requests issued before that, so let the run loop settle once.
static IOBluetoothDevice *device(const char *addr) {
	static BOOL ready;
	IOBluetoothDevice *dev = [IOBluetoothDevice deviceWithAddressString:[NSString stringWithUTF8String:addr]];
	if (!ready) {
		spinUntil([NSDate dateWithTimeIntervalSinceNow:0.5], ^{ return NO; });
		ready = YES;
	}
	return dev;
}

static int waitConnected(IOBluetoothDevice *dev, BOOL want, IOReturn r) {
	if (r != kIOReturnSuccess) return r;
	spinUntil([NSDate dateWithTimeIntervalSinceNow:10], ^{ return (BOOL)(dev.isConnected == want); });
	return dev.isConnected == want ? kIOReturnSuccess : kIOReturnTimeout;
}

static BOOL findChannel(IOBluetoothDevice *dev, BluetoothRFCOMMChannelID *out) {
	NSArray *spp = @[ [IOBluetoothSDPUUID uuid16:kBluetoothSDPUUID16ServiceClassSerialPort] ];
	for (IOBluetoothSDPServiceRecord *rec in dev.services) {
		if ([rec hasServiceFromArray:spp] && [rec getRFCOMMChannelID:out] == kIOReturnSuccess) return YES;
	}
	return NO;
}

int q30_open(const char *addr, void **handle) {
	@autoreleasepool {
		IOBluetoothDevice *dev = device(addr);
		if (dev == nil) return kIOReturnNotFound;
		if (!dev.isConnected) {
			IOReturn r = [dev openConnection];
			if (r != kIOReturnSuccess) return r;
		}
		BluetoothRFCOMMChannelID id;
		if (!findChannel(dev, &id)) {
			[dev performSDPQuery:nil];
			if (!findChannel(dev, &id)) return kIOReturnUnsupported;
		}
		Q30Channel *c = [Q30Channel new];
		c.buf = [NSMutableData data];
		IOBluetoothRFCOMMChannel *ch = nil;
		IOReturn r = [dev openRFCOMMChannelAsync:&ch withChannelID:id delegate:c];
		if (r != kIOReturnSuccess) return r;
		spinUntil([NSDate dateWithTimeIntervalSinceNow:10], ^{ return c.opened; });
		if (!c.opened) {
			[ch closeChannel];
			return kIOReturnTimeout;
		}
		if (c.openStatus != kIOReturnSuccess) return c.openStatus;
		c.channel = ch;
		*handle = (__bridge_retained void *)c;
		return kIOReturnSuccess;
	}
}

int q30_write(void *handle, const void *p, int n) {
	Q30Channel *c = (__bridge Q30Channel *)handle;
	return [c.channel writeSync:(void *)p length:(UInt16)n];
}

int q30_read(void *handle, void *out, int cap, double timeout) {
	@autoreleasepool {
		Q30Channel *c = (__bridge Q30Channel *)handle;
		spinUntil([NSDate dateWithTimeIntervalSinceNow:timeout], ^{ return (BOOL)(c.buf.length > 0 || c.closed); });
		if (c.buf.length == 0) return c.closed ? -1 : 0;
		int n = (int)MIN((NSUInteger)cap, c.buf.length);
		memcpy(out, c.buf.bytes, n);
		[c.buf replaceBytesInRange:NSMakeRange(0, n) withBytes:NULL length:0];
		return n;
	}
}

void q30_close(void *handle) {
	Q30Channel *c = (__bridge_transfer Q30Channel *)handle;
	[c.channel closeChannel];
}

int q30_connect(const char *addr) {
	IOBluetoothDevice *dev = device(addr);
	return dev == nil ? kIOReturnNotFound : waitConnected(dev, YES, [dev openConnection]);
}

int q30_disconnect(const char *addr) {
	IOBluetoothDevice *dev = device(addr);
	return dev == nil ? kIOReturnNotFound : waitConnected(dev, NO, [dev closeConnection]);
}

int q30_connected(const char *addr) {
	return device(addr).isConnected;
}

char *q30_find(const char *name) {
	@autoreleasepool {
		NSString *needle = [NSString stringWithUTF8String:name];
		for (IOBluetoothDevice *dev in [IOBluetoothDevice pairedDevices]) {
			if ([dev.name localizedCaseInsensitiveContainsString:needle]) {
				return strdup(dev.addressString.UTF8String);
			}
		}
		return NULL;
	}
}
