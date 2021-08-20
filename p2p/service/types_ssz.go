// Code generated by fastssz. DO NOT EDIT.
package service

import (
	ssz "github.com/ferranbt/fastssz"
)

// MarshalSSZ ssz marshals the DataMsgWrapper object
func (d *DataMsgWrapper) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(d)
}

// MarshalSSZTo ssz marshals the DataMsgWrapper object to a target array
func (d *DataMsgWrapper) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(17)

	// Field (0) 'Req'
	dst = ssz.MarshalBool(dst, d.Req)

	// Field (1) 'MsgType'
	dst = ssz.MarshalUint32(dst, d.MsgType)

	// Field (2) 'ReqID'
	dst = ssz.MarshalUint64(dst, d.ReqID)

	// Offset (3) 'Payload'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(d.Payload)

	// Field (3) 'Payload'
	if len(d.Payload) > 4096 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, d.Payload...)

	return
}

// UnmarshalSSZ ssz unmarshals the DataMsgWrapper object
func (d *DataMsgWrapper) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 17 {
		return ssz.ErrSize
	}

	tail := buf
	var o3 uint64

	// Field (0) 'Req'
	d.Req = ssz.UnmarshalBool(buf[0:1])

	// Field (1) 'MsgType'
	d.MsgType = ssz.UnmarshallUint32(buf[1:5])

	// Field (2) 'ReqID'
	d.ReqID = ssz.UnmarshallUint64(buf[5:13])

	// Offset (3) 'Payload'
	if o3 = ssz.ReadOffset(buf[13:17]); o3 > size {
		return ssz.ErrOffset
	}

	// Field (3) 'Payload'
	{
		buf = tail[o3:]
		if len(buf) > 4096 {
			return ssz.ErrBytesLength
		}
		if cap(d.Payload) == 0 {
			d.Payload = make([]byte, 0, len(buf))
		}
		d.Payload = append(d.Payload, buf...)
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the DataMsgWrapper object
func (d *DataMsgWrapper) SizeSSZ() (size int) {
	size = 17

	// Field (3) 'Payload'
	size += len(d.Payload)

	return
}