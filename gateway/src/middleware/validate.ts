import { Request, Response, NextFunction } from 'express'
import { ZodSchema, ZodError } from 'zod'

// Factory that returns an Express middleware validating req.body against a Zod schema.
// On failure returns 400 with a structured list of field errors.
export function validate(schema: ZodSchema) {
    return (req: Request, res: Response, next: NextFunction): void => {
        const result = schema.safeParse(req.body)

        if (!result.success) {
            const errors = (result.error as ZodError).errors.map((e) => ({
                field: e.path.join('.'),
                message: e.message,
            }))
            res.status(400).json({ error: 'Validation failed', details: errors })
            return
        }

        // Replace req.body with the parsed+coerced value so handlers
        // always receive clean typed data.
        req.body = result.data
        next()
    }
}