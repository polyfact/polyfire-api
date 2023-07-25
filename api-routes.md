# API Routes

## POST /generate

    -> Authentication: Required
    -> Args:
    	- task: STRING (in JSON body)
    	- memory_id: STRING (optional, in JSON body)
    -> Return: JSON of generated output or an error message.

## POST /memory

    -> Authentication: Required
    -> Args:
    	- id: UUID (in JSON body)
    	- input: STRING (in JSON body)
    -> Return: JSON response indicating success or an error message.

## PUT /memory

    -> Authentication: Required
    -> Args:
    	- id: UUID (in JSON body)
    	- input: STRING (in JSON body)
    -> Return: JSON response indicating success or an error message.

## GET /memories

    -> Authentication: Required
    -> Args: None
    -> Return: JSON array of user's memory IDs or an error message.
