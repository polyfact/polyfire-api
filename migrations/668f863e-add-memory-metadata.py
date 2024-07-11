def migrate(cur, rls=False):
    cur.execute("""
    ALTER TABLE embeddings ADD metadatas json DEFAULT '{}'::json;

    DROP FUNCTION retrieve_embeddings;
    CREATE OR REPLACE FUNCTION public.retrieve_embeddings(query_embedding vector, match_threshold double precision, match_count integer, memoryid uuid[], userid text)
     RETURNS TABLE(id uuid, content text, similarity double precision, metadatas json)
     LANGUAGE sql
     STABLE
    AS $function$
            SELECT
              embeddings.id,
              embeddings.content,
              1 - (embeddings.embedding <=> query_embedding) as similarity,
              embeddings.metadatas
            FROM embeddings
            JOIN memories ON embeddings.memory_id = memories.id
            WHERE
              1 - (embeddings.embedding <=> query_embedding) > match_threshold
              AND embeddings.memory_id = ANY(memoryid)
              AND (
                memories.user_id::text = userid
                OR memories.public = true
              )
            ORDER BY similarity DESC
            LIMIT match_count;
     $function$;
    """)

def rollback(cur, rls=False):
    cur.execute("""
    ALTER TABLE embeddings DROP COLUMN metadatas;

    DROP FUNCTION retrieve_embeddings;
    CREATE OR REPLACE FUNCTION public.retrieve_embeddings(query_embedding vector, match_threshold double precision, match_count integer, memoryid uuid[], userid text)
     RETURNS TABLE(id uuid, content text, similarity double precision)
     LANGUAGE sql
     STABLE
    AS $function$
            SELECT
              embeddings.id,
              embeddings.content,
              1 - (embeddings.embedding <=> query_embedding) as similarity
            FROM embeddings
            JOIN memories ON embeddings.memory_id = memories.id
            WHERE
              1 - (embeddings.embedding <=> query_embedding) > match_threshold
              AND embeddings.memory_id = ANY(memoryid)
              AND (
                memories.user_id::text = userid
                OR memories.public = true
              )
            ORDER BY similarity DESC
            LIMIT match_count;
     $function$;
    """)
