CREATE TEMP TABLE openrouter_models(
	model text,
	provider text,
	credit_input bigint,
	credit_type text,
	type text,
	credit_output bigint,
	image_url text,
	official_name text,
	hidden boolean,
	option_stream boolean,
	option_temperature boolean,
	option_stop boolean
);
\copy openrouter_models FROM 'codegen/openrouter-models.csv' DELIMITER ',' CSV HEADER ;
UPDATE models
SET
	credit_input = openrouter_models.credit_input,
  credit_output = openrouter_models.credit_output,
  hidden = openrouter_models.hidden,
  option_stream = openrouter_models.option_stream,
  option_temperature = openrouter_models.option_temperature,
  option_stop = openrouter_models.option_stop
FROM openrouter_models
WHERE models.provider = 'openrouter' AND models.model = openrouter_models.model;
DELETE FROM openrouter_models WHERE model IN (SELECT model FROM models WHERE provider = 'openrouter');
INSERT INTO models(
	model,
	provider,
	credit_input,
	credit_type,
	type,
	credit_output,
	image_url,
	official_name,
	hidden,
	option_stream,
	option_temperature,
	option_stop
) SELECT model, provider, credit_input, credit_type, type, credit_output, image_url, official_name, hidden, option_stream, option_temperature, option_stop FROM openrouter_models;
