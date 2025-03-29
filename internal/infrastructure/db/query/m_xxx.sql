select
	user_id,
	company_code,
	user_code,
    'X' || left(right(user_name, -1), - 1) || 'X' as user_name,
    'X' || left(right(user_name_kana, -1), - 1) || 'X' as user_name_kana,
from
	m_xxx