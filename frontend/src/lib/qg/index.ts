// Code generated by jtd-codegen for TypeScript v0.2.1

export type Qg = any;

/**
 * Error is returned on every API error.
 */
export interface Error {
  /**
   * Message is the error message
   */
  message: string;
}

/**
 * Game is the main game object. It contains all the information about the
 * game.
 */
export type Game = GameJeopardy | GameKahoot;

export interface GameJeopardy {
  game: "jeopardy";
  data: Jeopardy;
}

export interface GameKahoot {
  game: "kahoot";
  data: Kahoot;
}

/**
 * JeopardyGame is the game data for a Jeopardy game.
 */
export interface Jeopardy {
  categories: JeopardyCategory[];

  /**
   * moderators enables moderators being able to join.
   */
  moderators?: boolean;

  /**
   * require_name, if true, will require members to input a name before
   * we can participate.
   */
  require_name?: boolean;

  /**
   * score_multiplier is the score multiplier for each question. The
   * default is 100.
   */
  score_multiplier?: number;
}

/**
 * JeopardyCategory is a category in a Jeopardy game.
 */
export interface JeopardyCategory {
  /**
   * name is the name of the category.
   */
  name: string;

  /**
   * questions are the questions in the category.
   */
  questions: JeopardyQuestion[];
}

/**
 * JeopardyQuestion is a question in a Jeopardy game.
 */
export interface JeopardyQuestion {
  /**
   * answers are the possible answers.
   */
  answers: string[];

  /**
   * correct_answer is the correct answer within the list of answers
   * above. The index starts at 1.
   */
  correct_answer: number;

  /**
   * question is the question.
   */
  question: string;
}

/**
 * KahootGame is the game data for a Kahoot game.
 */
export interface Kahoot {
  /**
   * questions are the questions in the game.
   */
  questions: KahootQuestion[];

  /**
   * time_limit is the time limit for each question. The format is in
   * Go's time.Duration, e.g. 10s for 10 seconds.
   */
  time_limit: string;
}

/**
 * KahootQuestion is a question in a Kahoot game.
 */
export interface KahootQuestion {
  /**
   * answers are the possible answers.
   */
  answers: string[];

  /**
   * question is the question.
   */
  question: string;
}