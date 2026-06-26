import { Response } from 'express'
import * as grpc from '@grpc/grpc-js'

// Maps gRPC status codes to HTTP status codes.
// This is the mirror of errors.go on the Go side.
export function handleGrpcError(err: unknown, res: Response): void {
    if (!isGrpcError(err)) {
        res.status(500).json({ error: 'Internal server error' })
        return
    }

    switch (err.code) {
        case grpc.status.INVALID_ARGUMENT:
            res.status(400).json({ error: err.details })
            break
        case grpc.status.UNAUTHENTICATED:
            res.status(401).json({ error: err.details })
            break
        case grpc.status.PERMISSION_DENIED:
            res.status(403).json({ error: err.details })
            break
        case grpc.status.NOT_FOUND:
            res.status(404).json({ error: err.details })
            break
        case grpc.status.ALREADY_EXISTS:
            res.status(409).json({ error: err.details })
            break
        case grpc.status.RESOURCE_EXHAUSTED:
            res.status(429).json({ error: err.details })
            break
        default:
            res.status(500).json({ error: 'Internal server error' })
    }
}

interface GrpcError {
    code: grpc.status
    details: string
}

function isGrpcError(err: unknown): err is GrpcError {
    return (
        typeof err === 'object' &&
        err !== null &&
        'code' in err &&
        'details' in err
    )
}