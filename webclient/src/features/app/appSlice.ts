import { createAppSlice } from "../../app/createAppSlice";

export interface AppSliceState {
	readonly errors: string[];
}

const initialState: AppSliceState = {
	errors: []
};

export const appSlice = createAppSlice({
	name: "app",
	initialState,
	reducers: create => ({
		reportError: create.reducer<string>((state, action) => {
			state.errors = state.errors.concat(action.payload);
		})
	}),
	selectors: {
		selectErrors: app => app.errors
	}
});

export const { reportError } = appSlice.actions;
export const { selectErrors } = appSlice.selectors;
