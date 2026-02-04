/**
 * Augments @theme/Layout so that title and description (supported by
 * @docusaurus/theme-classic at runtime) are typed.
 */
declare module "@theme/Layout" {
  export interface Props {
    readonly title?: string;
    readonly description?: string;
  }
}
