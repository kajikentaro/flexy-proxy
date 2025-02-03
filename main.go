package main

import (
	"fmt"
	"regexp/syntax"
)

func IsRegexp(r *syntax.Regexp) bool {
	switch r.Op {
	case syntax.OpLiteral:
		return true
	case syntax.OpConcat:
		for _, sub := range r.Sub {
			if !IsRegexp(sub) {
				return false
			}
		}
		return true
	// (i.e.) contains a unsupported operator
	case syntax.OpEmptyMatch:
		return false
	default:
		return false
	}
}

// IsRegexCheck は、与えられた文字列が正規表現のメタ要素（演算子等）を含むかどうかを判定します。
// 戻り値：
//
//	第1引数(bool) → true: 正規表現としての要素を含む
//	                 false: 純粋な文字列のみ
//	第2引数(error) → 無効な正規表現の場合にエラーを返す。valid 前提なら呼び元の要件に合わせて扱う。
func IsRegexCheck(input string) (bool, error) {
	// Go の正規表現構文（Perl 互換）で parse
	re, err := syntax.Parse(input, syntax.Perl)
	if err != nil {
		// 無効な正規表現でパースに失敗した場合
		return false, err
	}

	// パース結果が純粋にリテラルのみかどうか
	if IsRegexp(re) {
		return false, nil
	}
	return true, nil
}

func main() {
	tests := []string{
		`\Qabc\E`,                  // → メタ要素なし (安全)
		`hello world hoge fuga \n`, // → メタ要素なし (安全)
		`hello*world`,              // → '*' (繰り返し) が含まれている
		`\bword\b`,                 // → \b など環境により追加されるメタ要素
		`^start`,                   // → '^' (アンカー)
		`end$`,                     // → '$'
		`(group)`,                  // → '( )'
		`abc[def]`,                 // → '[ ]'
		``,                         // → 空文字テスト
	}

	for _, t := range tests {
		isRegex, err := IsRegexCheck(t)
		if err != nil {
			fmt.Printf("Input: %q → パースエラー(無効な正規表現?): %v\n", t, err)
			continue
		}
		if isRegex {
			fmt.Printf("Input: %q → 正規表現としての機能 (NG)\n", t)
		} else {
			fmt.Printf("Input: %q → メタ要素無し (安全な文字列)\n", t)
		}
	}
}
