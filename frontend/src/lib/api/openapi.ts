/**
 * OpenAPI schema for kquiz
 * 0.0.1
 * DO NOT MODIFY - This file has been generated using oazapfts.
 * See https://www.npmjs.com/package/oazapfts
 */
import * as Oazapfts from "oazapfts/lib/runtime";
import * as QS from "oazapfts/lib/runtime/query";
export const defaults: Oazapfts.RequestOpts = {
    baseUrl: "/api/v0",
};
const oazapfts = Oazapfts.runtime(defaults);
export const servers = {
    version0ApiPath: "/api/v0"
};
export type GameType = "jeopardy" | "kahoot";
export type JeopardyQuestion = {
    question: string;
    answers: string[];
};
export type JeopardyCategory = {
    name: string;
    questions: JeopardyQuestion[];
};
export type JeopardyGame = {
    style: GameType;
    moderators?: boolean;
    member_name?: boolean;
    score_multiplier?: number;
    categories?: JeopardyCategory[];
};
export type KahootQuestion = {
    question: string;
    answers: string[];
};
export type KahootGame = {
    style: GameType;
    time_limit?: string;
    questions: KahootQuestion[];
};
export type GameData = {
    data?: ({
        style: "jeopardy";
    } & JeopardyGame) | ({
        style: "kahoot";
    } & KahootGame);
};
export type Error = {
    message: string;
};
/**
 * /ping returns pong
 */
export function ping(opts?: Oazapfts.RequestOpts) {
    return oazapfts.ok(oazapfts.fetchText("/ping", {
        ...opts
    }));
}
export function test(opts?: Oazapfts.RequestOpts) {
    return oazapfts.ok(oazapfts.fetchJson<{
        status: 200;
        data: GameData;
    } | {
        status: 400;
        data: Error;
    }>("/test", {
        ...opts
    }));
}
